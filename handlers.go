package vent

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"vent/auth"
	"vent/templates/gui"

	"entgo.io/ent/dialect/sql"
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
}

func NewHandler(secretProvider auth.SecretProvider) *Handler {
	return &Handler{
		mux:            http.NewServeMux(),
		authMiddleware: AuthMiddleware(secretProvider),
	}
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
		rows, err := schema.FilterFunc(order)
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

type FilterFunc func(order OrderOption) ([]DataRow, error)

type OrderOptionMap map[string]OrderOption
