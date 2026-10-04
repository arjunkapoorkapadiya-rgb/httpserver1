package main

import (
	"testing"
)

func TestRouterExactMatch(t *testing.T) {
	r := NewRouter()
	r.GET("/hello", func(req *Request) *Response {
		return NewResponse(200)
	})

	req := &Request{Method: "GET", Path: "/hello", Query: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 200 {
		t.Errorf("status: got %d, want 200", resp.Status)
	}
}

func TestRouterNotFound(t *testing.T) {
	r := NewRouter()
	r.GET("/hello", func(req *Request) *Response { return NewResponse(200) })

	req := &Request{Method: "GET", Path: "/nope", Query: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 404 {
		t.Errorf("status: got %d, want 404", resp.Status)
	}
}

func TestRouterPathParam(t *testing.T) {
	r := NewRouter()
	r.GET("/users/:id", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write(req.Query["id"])
		return resp
	})

	req := &Request{Method: "GET", Path: "/users/42", Query: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 200 {
		t.Fatalf("status: got %d", resp.Status)
	}
	if string(resp.Body) != "42" {
		t.Errorf("body: got %q, want %q", string(resp.Body), "42")
	}
}

func TestRouterWildcard(t *testing.T) {
	r := NewRouter()
	r.GET("/files/*path", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write(req.Query["path"])
		return resp
	})

	req := &Request{Method: "GET", Path: "/files/a/b/c.txt", Query: map[string]string{}}
	resp := r.Handle(req)
	if string(resp.Body) != "a/b/c.txt" {
		t.Errorf("body: got %q, want %q", string(resp.Body), "a/b/c.txt")
	}
}

func TestRouterMethodMismatch(t *testing.T) {
	r := NewRouter()
	r.GET("/hello", func(req *Request) *Response { return NewResponse(200) })

	req := &Request{Method: "POST", Path: "/hello", Query: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 404 {
		t.Errorf("status: got %d, want 404", resp.Status)
	}
}

func TestMiddlewareRuns(t *testing.T) {
	order := []string{}
	mw := func(next Handler) Handler {
		return func(req *Request) *Response {
			order = append(order, "before")
			resp := next(req)
			order = append(order, "after")
			return resp
		}
	}

	r := NewRouter()
	r.Use(mw)
	r.GET("/test", func(req *Request) *Response {
		order = append(order, "handler")
		return NewResponse(200)
	})

	req := &Request{Method: "GET", Path: "/test", Query: map[string]string{}}
	r.Handle(req)

	if len(order) != 3 {
		t.Fatalf("expected 3 events, got %d: %v", len(order), order)
	}
	if order[0] != "before" || order[1] != "handler" || order[2] != "after" {
		t.Errorf("wrong order: %v", order)
	}
}

func TestAuthMiddlewareBlocks(t *testing.T) {
	r := NewRouter()
	r.Use(AuthMiddleware("secret"))
	r.GET("/admin", func(req *Request) *Response {
		return NewResponse(200)
	})

	// Without token
	req := &Request{Method: "GET", Path: "/admin", Query: map[string]string{}, Headers: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 401 {
		t.Errorf("without token: got %d, want 401", resp.Status)
	}

	// With correct token
	req.Headers["Authorization"] = "Bearer secret"
	resp = r.Handle(req)
	if resp.Status != 200 {
		t.Errorf("with token: got %d, want 200", resp.Status)
	}

	// With wrong token
	req.Headers["Authorization"] = "Bearer wrong"
	resp = r.Handle(req)
	if resp.Status != 401 {
		t.Errorf("wrong token: got %d, want 401", resp.Status)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	r := NewRouter()
	r.Use(RecoveryMiddleware())
	r.GET("/panic", func(req *Request) *Response {
		panic("boom")
	})

	req := &Request{Method: "GET", Path: "/panic", Query: map[string]string{}}
	resp := r.Handle(req)
	if resp.Status != 500 {
		t.Errorf("status: got %d, want 500", resp.Status)
	}
}