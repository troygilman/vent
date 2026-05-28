package route

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

// Chain composes middleware so the first argument runs first on the request.
func Chain(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Router registers routes on a ServeMux tree with optional middleware inheritance.
type Router struct {
	mux        *http.ServeMux
	middleware []Middleware
}

// New returns a root router backed by a new ServeMux.
func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Use appends middleware inherited by this router and descendant groups/routes.
func (r *Router) Use(mw ...Middleware) {
	r.middleware = append(r.middleware, mw...)
}

// Mount registers a subtree at an absolute path prefix such as /admin/.
func (r *Router) Mount(absPrefix string, fn func(*Router), mws ...Middleware) error {
	mount, strip, err := NormalizeMountPrefix(absPrefix)
	if err != nil {
		return err
	}

	child := &Router{
		mux:        http.NewServeMux(),
		middleware: slices.Clone(r.middleware),
	}
	child.Use(mws...)
	fn(child)

	r.mux.Handle(mount, http.StripPrefix(strip, child.mux))
	return nil
}

// Group registers a relative path segment as a subtree. An empty segment applies
// middleware without an additional mount (middleware-only group).
func (r *Router) Group(segment string, fn func(*Router), mws ...Middleware) error {
	seg, err := NormalizeSegment(segment)
	if err != nil {
		return err
	}

	child := &Router{
		middleware: append(slices.Clone(r.middleware), mws...),
	}

	if seg == "" {
		child.mux = r.mux
		fn(child)
		return nil
	}

	child.mux = http.NewServeMux()
	fn(child)

	mount := "/" + seg + "/"
	r.mux.Handle(mount, http.StripPrefix("/"+seg, child.mux))
	return nil
}

// Handle registers a method-agnostic subtree handler.
func (r *Router) Handle(pattern string, h http.Handler, mws ...Middleware) error {
	pat, err := NormalizePattern(pattern)
	if err != nil {
		return err
	}
	handler := applyMiddleware(h, append(r.middleware, mws...))
	r.mux.Handle(pat, handler)
	return nil
}

// GET registers a GET handler on a normalized pattern.
func (r *Router) GET(pattern string, h http.Handler, mws ...Middleware) error {
	return r.method(http.MethodGet, pattern, h, mws...)
}

// POST registers a POST handler on a normalized pattern.
func (r *Router) POST(pattern string, h http.Handler, mws ...Middleware) error {
	return r.method(http.MethodPost, pattern, h, mws...)
}

// PUT registers a PUT handler on a normalized pattern.
func (r *Router) PUT(pattern string, h http.Handler, mws ...Middleware) error {
	return r.method(http.MethodPut, pattern, h, mws...)
}

// PATCH registers a PATCH handler on a normalized pattern.
func (r *Router) PATCH(pattern string, h http.Handler, mws ...Middleware) error {
	return r.method(http.MethodPatch, pattern, h, mws...)
}

// DELETE registers a DELETE handler on a normalized pattern.
func (r *Router) DELETE(pattern string, h http.Handler, mws ...Middleware) error {
	return r.method(http.MethodDelete, pattern, h, mws...)
}

func (r *Router) method(method, pattern string, h http.Handler, mws ...Middleware) error {
	pat, err := NormalizePattern(pattern)
	if err != nil {
		return err
	}
	handler := applyMiddleware(h, append(r.middleware, mws...))
	r.mux.Handle(method+" "+pat, handler)
	return nil
}

// Handler returns the router as an http.Handler.
func (r *Router) Handler() http.Handler {
	return r.mux
}

func applyMiddleware(h http.Handler, mws []Middleware) http.Handler {
	if len(mws) == 0 {
		return h
	}
	return Chain(mws...)(h)
}
