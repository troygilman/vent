package vent

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"vent/auth"
	"vent/templates/gui"

	"github.com/starfederation/datastar-go/datastar"
)

//go:embed static
var static embed.FS

func StaticDirHandler() http.Handler {
	return http.FileServerFS(static)
}

// HandlerConfig contains configuration options for the admin handler
type HandlerConfig struct {
	SecretProvider          auth.SecretProvider
	CredentialAuthenticator auth.CredentialAuthenticator
	AuthUserSchema          SchemaConfig
	AuthPasswordField       string // The field name containing the password hash (e.g., "password_hash")
	BasePath                string // Base path for admin routes (default: "/admin/")
	CookieName              string // Cookie name for auth token (default: "vent-auth-token")
}

// Handler is the main HTTP handler for the admin panel
type Handler struct {
	mux              *http.ServeMux
	config           HandlerConfig
	authMiddleware   Middleware
	userMiddleware   Middleware
	loggerMiddleware Middleware
	schemas          map[string]SchemaConfig
	tokenGenerator   auth.TokenGenerator
}

// NewHandler creates a new admin handler with the given configuration
func NewHandler(config HandlerConfig) *Handler {
	// Set defaults
	if config.BasePath == "" {
		config.BasePath = "/admin/"
	}
	if config.CookieName == "" {
		config.CookieName = "vent-auth-token"
	}

	h := &Handler{
		mux:              http.NewServeMux(),
		config:           config,
		schemas:          make(map[string]SchemaConfig),
		authMiddleware:   AuthentificationMiddleware(auth.NewJwtTokenAuthenticator(config.SecretProvider)),
		userMiddleware:   UserMiddleware(config.AuthUserSchema),
		loggerMiddleware: LoggerMiddleware(),
		tokenGenerator:   auth.NewJwtTokenGenerator(config.SecretProvider),
	}

	// Unauthorized Paths
	h.handle("GET "+config.BasePath+"static/", http.StripPrefix(config.BasePath, http.FileServerFS(static)), h.loggerMiddleware)
	h.handle("GET "+config.BasePath+"login/", h.getLoginHandler(), h.loggerMiddleware)
	h.handle("POST "+config.BasePath+"login/", h.postLoginHandler(), h.loggerMiddleware)

	// Authorized Paths
	h.handle("GET "+config.BasePath, h.authMiddleware(h.getAdminHandler()))

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// RegisterSchema registers a schema with the admin panel
func (h *Handler) RegisterSchema(schema SchemaConfig) {
	schema.Init()

	h.schemas[schema.Name] = schema

	path := h.config.BasePath + schema.Path()

	permission_suffix := strings.ToLower(schema.Name)

	h.handle("GET "+path, h.getSchemaListHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("view_"+permission_suffix))
	h.handle("POST "+path, h.postSchemaEntityHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("add_"+permission_suffix))
	h.handle("GET "+path+"add/", h.getSchemaEntityAddHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("add_"+permission_suffix))
	h.handle("GET "+path+"{id}/", h.getSchemaEntityHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("view_"+permission_suffix))
	h.handle("PATCH "+path+"{id}/", h.patchSchemaEntityHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("change_"+permission_suffix))
	h.handle("DELETE "+path+"{id}/", h.deleteSchemaEntityHandler(schema), h.loggerMiddleware, h.authMiddleware, h.userMiddleware, AuthorizationMiddleware("delete_"+permission_suffix))
}

func (h *Handler) handle(path string, handler http.Handler, middleware ...Middleware) {
	for _, m := range slices.Backward(middleware) {
		handler = m(handler)
	}
	h.mux.Handle(path, handler)
	log.Printf("Registered handler on %s", path)
}

// getLoginHandler returns the handler for GET /admin/login/
func (h *Handler) getLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		props := gui.LoginProps{}
		if err := gui.LoginPage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

// postLoginHandler returns the handler for POST /admin/login/
func (h *Handler) postLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signals struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			h.handleError(w, r, err, http.StatusBadRequest)
			return
		}

		loginProps := gui.LoginProps{}

		err := func() error {
			// Find user by email
			entities, err := h.config.AuthUserSchema.Client.List(r.Context(), ListOptions{
				Filters: map[string]any{"email": signals.Email},
				Limit:   1,
			})
			if err != nil {
				return err
			}
			if len(entities) != 1 {
				loginProps.EmailError = "User with this email does not exist"
				return errors.New("invalid email")
			}

			entity := entities[0]

			// Get password hash field
			passwordField, ok := entity.Get(h.config.AuthPasswordField)
			if !ok {
				return errors.New("corrupted user entity")
			}

			// Verify password
			if err := h.config.CredentialAuthenticator.Authenticate(signals.Password, passwordField.StringValue()); err != nil {
				loginProps.PasswordErrors = append(loginProps.PasswordErrors, "Password is invalid")
				return err
			}

			// Generate token
			claims := auth.NewClaims(entity.ID())
			token, err := h.tokenGenerator.Generate(claims)
			if err != nil {
				return err
			}

			// Set cookie
			http.SetCookie(w, &http.Cookie{
				Name:     h.config.CookieName,
				Value:    token,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})

			return nil
		}()

		sse := datastar.NewSSE(w, r)
		if err != nil {
			log.Println(err)
			if err := sse.PatchElementTempl(gui.LoginPage(loginProps)); err != nil {
				h.handleError(w, r, err, http.StatusInternalServerError)
			}
		} else {
			if err := sse.Redirect(h.config.BasePath); err != nil {
				h.handleError(w, r, err, http.StatusInternalServerError)
			}
		}
	})
}

// getAdminHandler returns the handler for GET /admin/
func (h *Handler) getAdminHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != h.config.BasePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		props := gui.AdminPageProps{
			LayoutProps: h.buildLayoutProps(""),
		}

		if err := gui.AdminPage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

// getSchemaListHandler returns the handler for GET /admin/{schema}s/
func (h *Handler) getSchemaListHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get list options from query params
		entities, err := schema.Client.List(r.Context(), ListOptions{
			OrderBy: "id",
		})
		if err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Build table props
		props := gui.SchemaTableProps{
			LayoutProps: h.buildLayoutProps(schema.Name),
			AdminPath:   h.config.BasePath,
			SchemaName:  schema.Name,
			Columns:     make([]gui.SchemaTableColumn, 0, len(schema.Columns)),
			Rows:        make([]gui.SchemaTableRow, 0, len(entities)),
		}

		for _, col := range schema.Columns {
			field := schema.LookupField(col)
			props.Columns = append(props.Columns, gui.SchemaTableColumn{
				Name:  field.Name,
				Label: field.Label,
				Type:  field.Type.String(),
			})
		}

		// Build rows
		for _, entity := range entities {
			row := gui.SchemaTableRow{
				Cells: make([]gui.SchemaTableCell, 0, len(schema.Columns)),
			}

			for i, col := range schema.Columns {
				field, ok := entity.Get(col)
				if !ok {
					row.Cells = append(row.Cells, gui.SchemaTableCell{Display: ""})
					continue
				}

				cell := gui.SchemaTableCell{
					Display: field.Display,
				}

				if i == 0 {
					cell.LinkURL = h.config.BasePath + schema.EntityPath(entity.ID())
				}

				// Add link URL for foreign key columns
				if field.Type == TypeForeignKey && field.Relation != nil {
					cell.LinkURL = fmt.Sprintf("%s%d/", field.Relation.TargetPath, field.Relation.TargetID)
				}

				row.Cells = append(row.Cells, cell)
			}

			props.Rows = append(props.Rows, row)
		}

		if err := gui.SchemaTablePage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

func (h *Handler) buildSchemaEntityFieldProps(ctx context.Context, schema SchemaConfig, entity EntityData) []gui.SchemaEntityFieldProps {
	fieldNames := []string{}
	if len(schema.FieldSets) > 0 {
		for _, fieldSet := range schema.FieldSets {
			fieldNames = append(fieldNames, fieldSet.Fields...)
		}
	} else {
		for _, field := range schema.Fields {
			fieldNames = append(fieldNames, field.Name)
		}
	}

	props := make([]gui.SchemaEntityFieldProps, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		field := schema.LookupField(fieldName)
		fieldProps := gui.SchemaEntityFieldProps{
			Name:     field.Name,
			Label:    field.Label,
			Type:     field.EffectiveInputType(),
			Value:    "",
			Editable: field.Editable,
		}

		var fieldValue FieldValue
		if entity != nil {
			var ok bool
			fieldValue, ok = entity.Get(field.Name)
			if ok {
				fieldProps.Value = fieldValue.Display
			}
		}

		// Add relation options if this is a foreign key
		if field.Type == TypeForeignKey && field.Relation != nil {
			foreignKeySchema, ok := h.schemas[field.Relation.TargetSchema]
			if !ok {
				panic("could not find foreign key schema with name " + field.Relation.TargetSchema)
			}
			foreignKeyOptions, err := foreignKeySchema.Client.List(ctx, QueryOptions{OrderBy: "id"})
			if err != nil {
				panic(err)
			}
			fieldProps.Options = make([]gui.SelectOption, len(foreignKeyOptions))
			ids := make(map[int]struct{})
			for _, relatedEntity := range fieldValue.RelationEntities() {
				ids[relatedEntity.ID()] = struct{}{}
			}
			for i, opt := range foreignKeyOptions {
				_, optionSelected := ids[opt.ID()]
				fieldProps.Options[i] = gui.SelectOption{
					Value:    opt.ID(),
					Label:    foreignKeySchema.EntityDisplayString(opt),
					Selected: optionSelected,
				}
			}
		}

		props = append(props, fieldProps)
	}
	return props
}

func (h *Handler) getSchemaEntityAddHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		props := gui.SchemaEntityAddProps{
			LayoutProps: h.buildLayoutProps(schema.Name),
			Fields:      h.buildSchemaEntityFieldProps(r.Context(), schema, nil),
			AdminPath:   h.config.BasePath,
			SchemaName:  schema.Name,
		}

		if err := gui.SchemaEntityAddPage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

// getSchemaEntityHandler returns the handler for GET /admin/{schema}s/{id}/
func (h *Handler) getSchemaEntityHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			h.handleError(w, r, fmt.Errorf("invalid id"), http.StatusBadRequest)
			return
		}

		entity, err := schema.Client.Get(r.Context(), id, GetOptions{
			WithEdges: schema.EdgeNames(),
		})
		if err != nil {
			h.handleError(w, r, err, http.StatusNotFound)
			return
		}

		// Build entity props
		props := gui.SchemaEntityProps{
			LayoutProps:   h.buildLayoutProps(schema.Name),
			Fields:        h.buildSchemaEntityFieldProps(r.Context(), schema, entity),
			AdminPath:     h.config.BasePath,
			SchemaName:    schema.Name,
			EntityID:      id,
			EntityDisplay: schema.EntityDisplayString(entity),
		}

		if err := gui.SchemaEntityPage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

func (h *Handler) parseEntityFormData(schema SchemaConfig, signals map[string]any) (map[string]any, error) {
	// Filter and parse fields
	data := make(map[string]any)
	for _, field := range schema.Fields {
		if !field.Editable {
			continue
		}
		fieldValue, ok := signals[field.Name]
		if !ok {
			continue
		}
		switch field.Type {
		case TypeForeignKey:
			anyIDs := fieldValue.([]any)
			intIDs := make([]int, 0, len(anyIDs))
			for _, anyID := range anyIDs {
				intID, err := strconv.Atoi(anyID.(string))
				if err != nil {
					return nil, err
				}
				intIDs = append(intIDs, intID)
			}
			data[field.Name] = intIDs
		case TypeForeignKeyUnique:
			id, err := strconv.Atoi(fieldValue.(string))
			if err != nil {
				return nil, err
			}
			data[field.Name] = id
		default:
			data[field.Name] = fieldValue
		}
	}

	// Apply field mappers (e.g., hash password, rename fields)
	if err := schema.ApplyFieldMappers(data); err != nil {
		return nil, fmt.Errorf("failed to apply field mappers: %w", err)
	}

	return data, nil
}

func (h *Handler) postSchemaEntityHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := make(map[string]any)
		if err := datastar.ReadSignals(r, &signals); err != nil {
			h.handleError(w, r, err, http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)

		data, err := h.parseEntityFormData(schema, signals)
		if err != nil {
			log.Printf("Error parsing entity data: %v", err)
			return
		}

		if _, err := schema.Client.Create(sse.Context(), data); err != nil {
			log.Printf("Error creating entity: %v", err)
			return
		}

		sse.Redirect(h.config.BasePath + schema.Path())
	})
}

// patchSchemaEntityHandler returns the handler for PATCH /admin/{schema}s/{id}/
func (h *Handler) patchSchemaEntityHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			h.handleError(w, r, fmt.Errorf("invalid id"), http.StatusBadRequest)
			return
		}

		signals := make(map[string]any)
		if err := datastar.ReadSignals(r, &signals); err != nil {
			h.handleError(w, r, err, http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)

		data, err := h.parseEntityFormData(schema, signals)
		if err != nil {
			log.Printf("Error parsing entity data: %v", err)
			return
		}

		if err := schema.Client.Update(sse.Context(), id, data); err != nil {
			log.Printf("Error updating entity: %v", err)
			// TODO: Send error via SSE
			return
		}

		sse.Redirect(h.config.BasePath + schema.Path())
	})
}

// deleteSchemaEntityHandler returns the handler for DELETE /admin/{schema}s/{id}/
func (h *Handler) deleteSchemaEntityHandler(schema SchemaConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			h.handleError(w, r, fmt.Errorf("invalid id"), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)

		if err := schema.Client.Delete(sse.Context(), id); err != nil {
			log.Printf("Error deleting entity: %v", err)
			// TODO: Send error via SSE
			return
		}

		sse.Redirect(h.config.BasePath + schema.Path())
	})
}

// buildLayoutProps creates LayoutProps with the current schemas
func (h *Handler) buildLayoutProps(activeSchemaName string) gui.LayoutProps {
	schemas := make([]gui.SchemaMetadata, 0, len(h.schemas))
	for _, schema := range h.schemas {
		schemas = append(schemas, gui.SchemaMetadata{
			Name: schema.Name,
			Path: h.config.BasePath + schema.Path(),
		})
	}
	slices.SortFunc(schemas, func(a, b gui.SchemaMetadata) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return gui.LayoutProps{
		Schemas:          schemas,
		ActiveSchemaName: activeSchemaName,
	}
}

// handleError logs the error and sends an appropriate HTTP response
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, status int) {
	log.Printf("Error [%s %s]: %v", r.Method, r.URL.Path, err)
	http.Error(w, err.Error(), status)
}
