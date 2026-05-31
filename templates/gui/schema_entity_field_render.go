package gui

import (
	"context"

	"github.com/a-h/templ"
)

func renderComponentHTML(ctx context.Context, component templ.Component) (string, error) {
	html, err := templ.ToGoHTML(ctx, component)
	if err != nil {
		return "", err
	}
	return string(html), nil
}

func RenderTextFieldHTML(ctx context.Context, props SchemaEntityTextFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityTextField(props))
}

func RenderPasswordFieldHTML(ctx context.Context, props SchemaEntityPasswordFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityPasswordField(props))
}

func RenderIntFieldHTML(ctx context.Context, props SchemaEntityIntFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityIntField(props))
}

func RenderFloatFieldHTML(ctx context.Context, props SchemaEntityFloatFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityFloatField(props))
}

func RenderBoolFieldHTML(ctx context.Context, props SchemaEntityBoolFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityBoolField(props))
}

func RenderTimeFieldHTML(ctx context.Context, props SchemaEntityTimeFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityTimeField(props))
}

func RenderForeignKeyUniqueFieldHTML(ctx context.Context, props SchemaEntityForeignKeyUniqueFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityForeignKeyUniqueField(props))
}

func RenderForeignKeyFieldHTML(ctx context.Context, props SchemaEntityForeignKeyFieldProps) (string, error) {
	return renderComponentHTML(ctx, SchemaEntityForeignKeyField(props))
}
