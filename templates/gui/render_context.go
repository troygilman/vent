package gui

import "context"

// RenderContext carries request-scoped flags that influence admin UI rendering.
// It is intentionally extendable — add fields here as new render concerns appear.
type RenderContext struct {
	CanUpdate bool
	CanDelete bool
}

type renderContextKey struct{}

// WithRenderContext attaches rc to ctx for field renderers and page templates.
func WithRenderContext(ctx context.Context, rc RenderContext) context.Context {
	return context.WithValue(ctx, renderContextKey{}, rc)
}

// RenderContextFrom returns the RenderContext stored on ctx, if any.
func RenderContextFrom(ctx context.Context) (RenderContext, bool) {
	rc, ok := ctx.Value(renderContextKey{}).(RenderContext)
	return rc, ok
}
