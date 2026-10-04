package main

import (
	"strings"
)

// Handler is a function that handles a request and returns a response.
type Handler func(*Request) *Response

// Route is one registered route.
type Route struct {
	Method   string
	Segments []string
	Handler  Handler
}

// Router matches requests to handlers.
type Router struct {
	routes      []Route
	middlewares []Middleware
}

// NewRouter creates an empty router.
func NewRouter() *Router {
	return &Router{}
}

// Use adds a global middleware that applies to all routes.
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Register adds a route.
func (r *Router) register(method, pattern string, h Handler) {
	pattern = strings.Trim(pattern, "/")
	var segs []string
	if pattern != "" {
		segs = strings.Split(pattern, "/")
	}
	r.routes = append(r.routes, Route{
		Method:   method,
		Segments: segs,
		Handler:  h,
	})
}

func (r *Router) GET(pattern string, h Handler)    { r.register("GET", pattern, h) }
func (r *Router) POST(pattern string, h Handler)   { r.register("POST", pattern, h) }
func (r *Router) PUT(pattern string, h Handler)    { r.register("PUT", pattern, h) }
func (r *Router) DELETE(pattern string, h Handler) { r.register("DELETE", pattern, h) }

// Handle finds a matching route, wraps with middleware, and calls the handler.
func (r *Router) Handle(req *Request) *Response {
	reqPath := strings.Trim(req.Path, "/")
	var segments []string
	if reqPath != "" {
		segments = strings.Split(reqPath, "/")
	}

	// Find matching route
	var matchedHandler Handler
	for _, rt := range r.routes {
		if rt.Method != req.Method {
			continue
		}
		params, ok := matchRoute(rt.Segments, segments)
		if !ok {
			continue
		}
		for k, v := range params {
			req.Query[k] = v
		}
		matchedHandler = rt.Handler
		break
	}

	// If no route matched, use a default 404 handler
	if matchedHandler == nil {
		matchedHandler = func(req *Request) *Response {
			resp := NewResponse(404)
			resp.Write("Not Found: " + req.Path + "\n")
			return resp
		}
	}

	// Apply middlewares (in reverse so first-added is outermost)
	wrapped := matchedHandler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		wrapped = r.middlewares[i](wrapped)
	}

	return wrapped(req)
}

// matchRoute matches segments against pattern segments.
func matchRoute(pattern, segments []string) (map[string]string, bool) {
	params := map[string]string{}

	wildcardIdx := -1
	for i, seg := range pattern {
		if strings.HasPrefix(seg, "*") {
			wildcardIdx = i
			break
		}
	}

	if wildcardIdx >= 0 {
		if len(segments) < wildcardIdx {
			return nil, false
		}
		for i := 0; i < wildcardIdx; i++ {
			if !matchSegment(pattern[i], segments[i], params) {
				return nil, false
			}
		}
		name := pattern[wildcardIdx][1:]
		rest := strings.Join(segments[wildcardIdx:], "/")
		params[name] = rest
		return params, true
	}

	if len(pattern) != len(segments) {
		return nil, false
	}
	for i := range pattern {
		if !matchSegment(pattern[i], segments[i], params) {
			return nil, false
		}
	}
	return params, true
}

func matchSegment(pat, seg string, params map[string]string) bool {
	if strings.HasPrefix(pat, ":") {
		params[pat[1:]] = seg
		return true
	}
	return pat == seg
}