package vent

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"vent/auth"
	"vent/templates/gui"

	"entgo.io/ent/dialect/sql"
	"github.com/starfederation/datastar-go/datastar"
)

//go:embed static
var static embed.FS

func StaticDirHandler() http.Handler {
	return http.FileServerFS(static)
}

type Handler struct {
	mux            *http.ServeMux
	authMiddleware Middleware
	schemas        []gui.SchemaMetadata
	tokenGenerator auth.TokenGenerator
	loginHandler   LoginHandler
}

func NewHandler(
	secretProvider auth.SecretProvider,
	loginHandler LoginHandler,
) *Handler {
	handler := &Handler{
		mux:            http.NewServeMux(),
		authMiddleware: NewAuthMiddleware(secretProvider),
		tokenGenerator: auth.NewJwtTokenGenerator(secretProvider),
		loginHandler:   loginHandler,
	}

	// Unauthorized Paths
	handler.mux.Handle("GET /admin/static/", http.StripPrefix("/admin/", http.FileServerFS(static)))
	handler.mux.Handle("GET /admin/login/", handler.getAdminLoginHandler())
	handler.mux.Handle("POST /admin/login/", handler.postAdminLoginHandler())

	// Authorized Paths
	handler.mux.Handle("GET /admin/", handler.authMiddleware(handler.getAdminHandler()))

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

type UserCredential struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (handler *Handler) postAdminLoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signals UserCredential
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := func() error {
			userID, err := handler.loginHandler(r.Context(), signals)
			if err != nil {
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
		Path: fmt.Sprintf("/admin/%s/", strings.ToLower(schema.Name)),
	}
	handler.schemas = append(handler.schemas, metadata)
	handler.mux.Handle(fmt.Sprintf("GET %s", metadata.Path), handler.authMiddleware(handler.getSchemaTableHandler(schema)))
}

func (handler *Handler) getSchemaTableHandler(schema SchemaParams) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order := schema.OrderOptionMap["id"]
		rows, err := schema.FilterFunc(r.Context(), order)
		if err != nil {
			panic(err)
		}

		props := gui.SchemaTableProps{
			LayoutProps: gui.LayoutProps{
				Schemas:          handler.schemas,
				ActiveSchemaName: schema.Name,
			},
			Columns:    schema.Columns,
			Rows:       rows,
			AdminPath:  "/admin/",
			SchemaName: schema.Name,
		}

		gui.SchemaTablePage(props).Render(r.Context(), w)
	})
}

type SchemaParams struct {
	Name           string
	FilterFunc     FilterFunc
	OrderOptionMap OrderOptionMap
	Columns        []gui.SchemaTableColumn
}

type DataRow = []string

type OrderOption = func(*sql.Selector)

type FilterFunc func(ctx context.Context, order OrderOption) ([]DataRow, error)

type OrderOptionMap map[string]OrderOption

type LoginHandler func(ctx context.Context, credential UserCredential) (id int, err error)
