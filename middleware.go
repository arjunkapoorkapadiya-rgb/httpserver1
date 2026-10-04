package main

import (
	"fmt"
	"strings"
	"time"
)

// Middleware wraps a handler with extra behavior.
type Middleware func(Handler) Handler

// Chain applies middlewares in order. The first middleware runs first.
func Chain(mws ...Middleware) Middleware {
	return func(next Handler) Handler {
		// Apply in reverse so first middleware is outermost
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// LoggingMiddleware prints each request + response time.
func LoggingMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			start := time.Now()
			resp := next(req)
			dur := time.Since(start)

			// Pretty-print a log line
			coloredStatus := fmt.Sprintf("%d", resp.Status)
			fmt.Printf("  → %s %s  [%s]  %s\n",
				req.Method, req.Path, coloredStatus, dur.Round(time.Microsecond))
			return resp
		}
	}
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers.
func CORSMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			resp := next(req)
			resp.SetHeader("Access-Control-Allow-Origin", "*")
			resp.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			resp.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return resp
		}
	}
}

// AuthMiddleware requires "Authorization: Bearer <token>".
// If the token doesn't match, returns 401 without calling the handler.
func AuthMiddleware(token string) Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			authHeader := req.Headers["Authorization"]
			expected := "Bearer " + token
			if authHeader != expected {
				resp := NewResponse(401)
				resp.Write("Unauthorized\n")
				return resp
			}
			return next(req)
		}
	}
}

// RecoveryMiddleware catches panics in handlers and returns 500.
func RecoveryMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) (resp *Response) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  ⚠ panic recovered: %v\n", r)
					resp = NewResponse(500)
					resp.Write("Internal Server Error\n")
				}
			}()
			return next(req)
		}
	}
}

// MethodOverrideMiddleware lets clients send POST with header
// "X-HTTP-Method-Override: DELETE" to route to a DELETE handler.
// (Useful for browsers that only support GET/POST.)
func MethodOverrideMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(req *Request) *Response {
			if override := req.Headers["X-HTTP-Method-Override"]; override != "" {
				req.Method = strings.ToUpper(override)
			}
			return next(req)
		}
	}
}