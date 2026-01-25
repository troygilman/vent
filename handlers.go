package vent

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"iter"
	"log"
	"net/http"
	"strconv"
	"strings"
	"vent/auth"
	"vent/templates/gui"
	"vent/utils"

	"entgo.io/ent"

	"entgo.io/ent/dialect/sql"
	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

//go:embed static
var static embed.FS

func StaticDirHandler() http.Handler {
	return http.FileServerFS(static)
}

type Handler struct {
	mux                     *http.ServeMux
	authMiddleware          Middleware
	schemas                 []gui.SchemaMetadata
	tokenGenerator          auth.TokenGenerator
	credentialAuthenticator auth.CredentialAuthenticator
}

func NewHandler(
	secretProvider auth.SecretProvider,
	credentialAuthenticator auth.CredentialAuthenticator,
	authUserSchema SchemaParams,
) *Handler {
	handler := &Handler{
		mux:                     http.NewServeMux(),
		authMiddleware:          NewAuthMiddleware(secretProvider),
		tokenGenerator:          auth.NewJwtTokenGenerator(secretProvider),
		credentialAuthenticator: credentialAuthenticator,
	}

	// Unauthorized Paths
	handler.handle("GET /admin/static/", http.StripPrefix("/admin/", http.FileServerFS(static)))
	handler.handle("GET /admin/login/", handler.getAdminLoginHandler())
	handler.handle("POST /admin/login/", handler.postAdminLoginHandler(authUserSchema))

	// Authorized Paths
	handler.handle("GET /admin/", handler.authMiddleware(handler.getAdminHandler()))

	return handler
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mux.ServeHTTP(w, r)
}

func (handler *Handler) getAdminLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		props := gui.LoginProps{}
		gui.LoginPage(props).Render(r.Context(), w)
	})
}

func (handler *Handler) postAdminLoginHandler(schema SchemaParams) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signals struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := func() error {
			entity, err := schema.SchemaClient.QueryOnly(r.Context(), []Predicate{sql.FieldEQ("email", signals.Email)})
			userID := entity.Get("id").(int)
			passwordHash := entity.Get("password_hash").(string)

			if err := handler.credentialAuthenticator.Authenticate(signals.Password, passwordHash); err != nil {
				return err
			}

			claims := auth.NewClaims(userID)
			token, err := handler.tokenGenerator.Generate(claims)
			if err != nil {
				return err
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "vent-auth-token",
				Value:    token,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})

			sse := datastar.NewSSE(w, r)
			sse.Redirect("/admin/")
			return nil
		}()

		if err != nil {
			log.Println(err.Error())
			// sse := datastar.NewSSE(w, r)
			// if IsNotFound(err) {
			// 	sse.PatchElementTempl(gui.Login(gui.LoginProps{
			// 		EmailError: "User with email not found",
			// 	}))
			// } else if errors.Is(err, vent.ErrPasswordMismatch) {
			// 	sse.PatchElementTempl(gui.Login(gui.LoginProps{
			// 		PasswordErrors: []string{
			// 			"Invalid password",
			// 		},
			// 	}))
			// }
		}
	})
}

func (handler *Handler) getAdminHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		props := gui.AdminPageProps{
			LayoutProps: gui.LayoutProps{
				Schemas: handler.schemas,
			},
		}

		gui.AdminPage(props).Render(r.Context(), w)
	})
}

func (handler *Handler) RegisterSchema(schema SchemaParams) {
	metadata := gui.SchemaMetadata{
		Name: schema.Name,
		Path: fmt.Sprintf("/admin/%ss/", strings.ToLower(schema.Name)),
	}
	handler.schemas = append(handler.schemas, metadata)
	handler.handle(fmt.Sprintf("GET %s", metadata.Path), handler.authMiddleware(handler.getSchemaTableHandler(schema)))
	handler.handle(fmt.Sprintf("GET %s{id}/", metadata.Path), handler.authMiddleware(handler.getSchemaEntityHandler(schema)))
	handler.handle(fmt.Sprintf("POST %s{id}/", metadata.Path), handler.authMiddleware(handler.postSchemaEntityHandler(schema, metadata)))
	handler.handle(fmt.Sprintf("DELETE %s{id}/", metadata.Path), handler.authMiddleware(handler.deleteSchemaEntityHandler(schema, metadata)))
}

func (handler *Handler) handle(path string, h http.Handler) {
	handler.mux.Handle(path, h)
	log.Printf("Registered handler on %s", path)
}

func (handler *Handler) getSchemaTableHandler(schema SchemaParams) http.Handler {
	fieldMap := make(map[string]SchemaColumn)
	for _, column := range schema.Columns {
		fieldMap[column.Name] = column
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		column, ok := fieldMap["id"]
		if !ok {
			panic("field does not exist")
		}

		entities, err := schema.SchemaClient.Query(ctx, nil, []OrderOption{sql.OrderByField(column.Name).ToFunc()})
		if err != nil {
			panic(err)
		}

		props := gui.SchemaTableProps{
			LayoutProps: gui.LayoutProps{
				Schemas:          handler.schemas,
				ActiveSchemaName: schema.Name,
			},
			Columns:    make([]gui.SchemaTableColumn, len(schema.Columns)),
			Rows:       []DataRow{},
			AdminPath:  "/admin/",
			SchemaName: schema.Name,
		}

		for idx, column := range schema.Columns {
			props.Columns[idx] = gui.SchemaTableColumn{
				Name:  column.Name,
				Label: column.Label,
				Type:  column.Type,
			}
		}

		for entity := range entities {
			row := make(DataRow, len(schema.Columns))
			for columnIdx, column := range schema.Columns {
				row[columnIdx] = utils.Stringify(entity.Get(column.Name), column.Type)
			}
			props.Rows = append(props.Rows, row)
		}

		gui.SchemaTablePage(props).Render(r.Context(), w)
	})
}

func (handler *Handler) getSchemaEntityHandler(schema SchemaParams) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			panic(err)
		}

		component, err := handler.buildSchemaEntityComponent(ctx, schema, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := component.Render(ctx, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (handler *Handler) postSchemaEntityHandler(schema SchemaParams, metadata gui.SchemaMetadata) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signals := make(map[string]any)
		if err := datastar.ReadSignals(r, &signals); err != nil {
			panic(err)
		}

		sse := datastar.NewSSE(w, r)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			panic(err)
		}

		mutations := []Mutation{}
		for fieldName, value := range signals {
			mutations = append(mutations, Mutation{
				FieldName: fieldName,
				Value:     value,
			})
		}

		err = schema.SchemaClient.Update(sse.Context(), []Predicate{sql.FieldEQ("id", id)}, mutations)
		if err != nil {
			panic(err)
		}

		sse.Redirect(metadata.Path)
	})
}

func (handler *Handler) deleteSchemaEntityHandler(schema SchemaParams, metadata gui.SchemaMetadata) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			panic(err)
		}

		count, err := schema.SchemaClient.Delete(sse.Context(), []Predicate{sql.FieldEQ("id", id)})
		if err != nil {
			panic(err)
		}

		if count == 0 {
			panic("could not delete entity")
		}

		if count > 1 {
			panic("deleted more than 1 entity")
		}

		sse.Redirect(metadata.Path)
	})
}

func (handler *Handler) getEntityById(ctx context.Context, schema SchemaParams, id int) (Entity, error) {
	entities, err := schema.SchemaClient.Query(ctx, []Predicate{sql.FieldEQ("id", id)}, []OrderOption{})
	if err != nil {
		return nil, err
	}
	for entity := range entities {
		return entity, nil
	}
	return nil, errors.New("could not find entity")
}

func (handler *Handler) buildSchemaEntityComponent(ctx context.Context, schema SchemaParams, id int) (templ.Component, error) {
	entity, err := handler.getEntityById(ctx, schema, id)
	if err != nil {
		return nil, err
	}

	props := gui.SchemaEntityProps{
		LayoutProps: gui.LayoutProps{
			Schemas:          handler.schemas,
			ActiveSchemaName: schema.Name,
		},
		Fields:     make([]gui.SchemaEntityFieldProps, len(schema.Columns)),
		AdminPath:  "/admin/",
		SchemaName: schema.Name,
		EntityID:   id,
	}

	for idx, column := range schema.Columns {
		props.Fields[idx] = gui.SchemaEntityFieldProps{
			Name:  column.Name,
			Label: column.Label,
			Type:  column.Type,
			Value: utils.Stringify(entity.Get(column.Name), column.Type),
		}
	}

	return gui.SchemaEntityPage(props), nil
}

type SchemaParams struct {
	Name         string
	Columns      []SchemaColumn
	SchemaClient SchemaClient
}

type SchemaColumn struct {
	Name  string
	Label string
	Type  string
}

type DataRow = []string

type OrderOption = func(*sql.Selector)

type Predicate = func(*sql.Selector)

type Entity interface {
	Get(field string) ent.Value
}

type SchemaClient interface {
	Query(ctx context.Context, predicates []Predicate, orders []OrderOption) (iter.Seq[Entity], error)
	QueryOnly(ctx context.Context, predicates []Predicate) (Entity, error)
	Update(ctx context.Context, predicates []Predicate, mutations []Mutation) error
	Delete(ctx context.Context, predicates []Predicate) (int, error)
}

type Mutation struct {
	FieldName string
	Value     any
}
