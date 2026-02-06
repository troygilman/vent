package vent

import (
	"embed"
	"fmt"
	"log"
	"net/http"
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
	mux            *http.ServeMux
	config         HandlerConfig
	authMiddleware Middleware
	userMiddleware Middleware
	schemas        []SchemaConfig
	tokenGenerator auth.TokenGenerator
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

	handler := &Handler{
		mux:            http.NewServeMux(),
		config:         config,
		authMiddleware: AuthentificationMiddleware(auth.NewJwtTokenAuthenticator(config.SecretProvider)),
		userMiddleware: UserMiddleware(config.AuthUserSchema),
		tokenGenerator: auth.NewJwtTokenGenerator(config.SecretProvider),
	}

	// Unauthorized Paths
	handler.handle("GET "+config.BasePath+"static/", http.StripPrefix(config.BasePath, http.FileServerFS(static)))
	handler.handle("GET "+config.BasePath+"login/", handler.getLoginHandler())
	handler.handle("POST "+config.BasePath+"login/", handler.postLoginHandler())

	// Authorized Paths
	handler.handle("GET "+config.BasePath, handler.authMiddleware(handler.getAdminHandler()))

	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// RegisterSchema registers a schema with the admin panel
func (h *Handler) RegisterSchema(schema SchemaConfig) {
	// Set the admin path if not already set
	if schema.AdminPath == "" {
		schema.AdminPath = h.config.BasePath
	}

	h.schemas = append(h.schemas, schema)

	path := schema.Path()

	permission_suffix := strings.ToLower(schema.Name)

	h.handle("GET "+path, h.authMiddleware(h.userMiddleware(AuthorizationMiddleware("view_"+permission_suffix)(h.getSchemaListHandler(schema)))))
	h.handle("GET "+path+"{id}/", h.authMiddleware(h.userMiddleware(AuthorizationMiddleware("view_"+permission_suffix)(h.getSchemaEntityHandler(schema)))))
	// h.handle("POST "+path+"{id}/", h.authMiddleware(h.userMiddleware(AuthorizationMiddleware("add_"+permission_suffix)(h.postSchemaEntityHandler(schema)))))
	h.handle("PATCH "+path+"{id}/", h.authMiddleware(h.userMiddleware(AuthorizationMiddleware("change_"+permission_suffix)(h.patchSchemaEntityHandler(schema)))))
	h.handle("DELETE "+path+"{id}/", h.authMiddleware(h.userMiddleware(AuthorizationMiddleware("delete_"+permission_suffix)(h.deleteSchemaEntityHandler(schema)))))
}

func (h *Handler) handle(path string, handler http.Handler) {
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

		// Find user by email
		entities, err := h.config.AuthUserSchema.Client.List(r.Context(), ListOptions{
			Filters: map[string]any{"email": signals.Email},
			Limit:   1,
		})
		if err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
			return
		}
		if len(entities) == 0 {
			h.handleError(w, r, fmt.Errorf("user not found"), http.StatusUnauthorized)
			return
		}

		entity := entities[0]

		// Get password hash field
		passwordField, ok := entity.Get(h.config.AuthPasswordField)
		if !ok {
			h.handleError(w, r, fmt.Errorf("password field not found"), http.StatusInternalServerError)
			return
		}

		// Verify password
		if err := h.config.CredentialAuthenticator.Authenticate(signals.Password, passwordField.StringValue()); err != nil {
			h.handleError(w, r, fmt.Errorf("invalid password"), http.StatusUnauthorized)
			return
		}

		// Generate token
		claims := auth.NewClaims(entity.ID())
		token, err := h.tokenGenerator.Generate(claims)
		if err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     h.config.CookieName,
			Value:    token,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to admin home
		sse := datastar.NewSSE(w, r)
		sse.Redirect(h.config.BasePath)
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
		opts := ListOptions{
			OrderBy: "id",
		}

		entities, err := schema.Client.List(r.Context(), opts)
		if err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Build table props
		props := gui.SchemaTableProps{
			LayoutProps: h.buildLayoutProps(schema.Name),
			Columns:     make([]gui.SchemaTableColumn, len(schema.Columns)),
			Rows:        make([]gui.SchemaTableRow, 0, len(entities)),
		}

		// Build column headers
		for i, col := range schema.Columns {
			props.Columns[i] = gui.SchemaTableColumn{
				Name:  col.Name,
				Label: col.Label,
				Type:  col.Type.String(),
			}
		}

		// Build rows
		for _, entity := range entities {
			row := gui.SchemaTableRow{
				Values: make([]gui.SchemaTableCell, len(schema.Columns)),
			}

			for i, col := range schema.Columns {
				field, ok := entity.Get(col.Name)
				if !ok {
					row.Values[i] = gui.SchemaTableCell{Display: ""}
					continue
				}

				cell := gui.SchemaTableCell{
					Display: field.Display,
				}

				// Add link URL for ID column (link to entity detail page)
				if col.Name == "id" {
					cell.LinkURL = fmt.Sprintf("%s%d/", schema.Path(), entity.ID())
				}

				// Add link URL for foreign key columns
				if field.Type == TypeForeignKey && field.Relation != nil {
					cell.LinkURL = fmt.Sprintf("%s%d/", field.Relation.TargetPath, field.Relation.TargetID)
				}

				row.Values[i] = cell
			}

			props.Rows = append(props.Rows, row)
		}

		if err := gui.SchemaTablePage(props).Render(r.Context(), w); err != nil {
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

		entity, err := schema.Client.Get(r.Context(), id)
		if err != nil {
			h.handleError(w, r, err, http.StatusNotFound)
			return
		}

		// Build entity props
		props := gui.SchemaEntityProps{
			LayoutProps: h.buildLayoutProps(schema.Name),
			Fields:      make([]gui.SchemaEntityFieldProps, 0, len(schema.Columns)),
			AdminPath:   h.config.BasePath,
			SchemaName:  schema.Name,
			EntityID:    id,
		}

		for _, col := range schema.Columns {
			field, ok := entity.Get(col.Name)
			if !ok {
				continue
			}

			fieldProps := gui.SchemaEntityFieldProps{
				Name:     col.Name,
				Label:    col.Label,
				Type:     col.Type.String(),
				Value:    field.Display,
				Editable: col.Editable,
			}

			// Add relation options if this is a foreign key
			if col.Type == TypeForeignKey && col.Relation != nil {
				options, err := schema.Client.GetRelationOptions(r.Context(), col.Relation)
				if err != nil {
					log.Printf("Error getting relation options for %s: %v", col.Name, err)
				} else {
					fieldProps.Options = make([]gui.SelectOption, len(options))
					for i, opt := range options {
						fieldProps.Options[i] = gui.SelectOption{
							Value:    opt.Value,
							Label:    opt.Label,
							Selected: field.Type == TypeForeignKey && field.IntValue() == opt.Value,
						}
					}
				}
			}

			props.Fields = append(props.Fields, fieldProps)
		}

		if err := gui.SchemaEntityPage(props).Render(r.Context(), w); err != nil {
			h.handleError(w, r, err, http.StatusInternalServerError)
		}
	})
}

// patchSchemaEntityHandler returns the handler for POST /admin/{schema}s/{id}/
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

		// Filter to only editable fields
		data := make(map[string]any)
		for _, col := range schema.Columns {
			if col.Editable {
				if val, ok := signals[col.Name]; ok {
					data[col.Name] = val
				}
			}
		}

		if err := schema.Client.Update(sse.Context(), id, data); err != nil {
			log.Printf("Error updating entity: %v", err)
			// TODO: Send error via SSE
			return
		}

		sse.Redirect(schema.Path())
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

		sse.Redirect(schema.Path())
	})
}

// buildLayoutProps creates LayoutProps with the current schemas
func (h *Handler) buildLayoutProps(activeSchemaName string) gui.LayoutProps {
	schemas := make([]gui.SchemaMetadata, len(h.schemas))
	for i, s := range h.schemas {
		schemas[i] = gui.SchemaMetadata{
			Name: s.Name,
			Path: s.Path(),
		}
	}
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
