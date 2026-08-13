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

// applyFormEditable ANDs a field's own Editable flag with render-context update
// permission. When no RenderContext is set (e.g. create forms), the field flag wins.
func applyFormEditable(ctx context.Context, editable bool) bool {
	if rc, ok := RenderContextFrom(ctx); ok {
		return editable && rc.CanUpdate
	}
	return editable
}

func RenderTextFieldHTML(ctx context.Context, props SchemaEntityTextFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	if rc, ok := RenderContextFrom(ctx); ok && !rc.CanUpdate {
		props.ActionLabel = ""
		props.ActionURL = ""
	}
	return renderComponentHTML(ctx, SchemaEntityTextField(props))
}

func RenderPasswordFieldHTML(ctx context.Context, props SchemaEntityPasswordFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityPasswordField(props))
}

func RenderIntFieldHTML(ctx context.Context, props SchemaEntityIntFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityIntField(props))
}

func RenderFloatFieldHTML(ctx context.Context, props SchemaEntityFloatFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityFloatField(props))
}

func RenderBoolFieldHTML(ctx context.Context, props SchemaEntityBoolFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityBoolField(props))
}

func RenderTimeFieldHTML(ctx context.Context, props SchemaEntityTimeFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityTimeField(props))
}

func RenderForeignKeyUniqueFieldHTML(ctx context.Context, props SchemaEntityForeignKeyUniqueFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityForeignKeyUniqueField(props))
}

func RenderForeignKeyFieldHTML(ctx context.Context, props SchemaEntityForeignKeyFieldProps) (string, error) {
	props.Editable = applyFormEditable(ctx, props.Editable)
	return renderComponentHTML(ctx, SchemaEntityForeignKeyField(props))
}
