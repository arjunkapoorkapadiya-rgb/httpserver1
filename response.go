package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Response is what we'll send back.
type Response struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

// NewResponse creates a response with a status code.
func NewResponse(status int) *Response {
	return &Response{
		Status:  status,
		Headers: map[string]string{},
	}
}

// SetHeader sets a response header.
func (r *Response) SetHeader(key, val string) {
	r.Headers[canonicalHeader(key)] = val
}

// Write sets the body from a string.
func (r *Response) Write(s string) {
	r.Body = []byte(s)
}

// WriteBytes sets the body from bytes.
func (r *Response) WriteBytes(b []byte) {
	r.Body = b
}

// statusText maps status codes to reason phrases.
func statusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Found"
	case 304:
		return "Not Modified"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 500:
		return "Internal Server Error"
	case 501:
		return "Not Implemented"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	}
	return "Unknown"
}

// write serializes the response to the writer.
func (r *Response) write(w *bufio.Writer) error {
	// Status line
	fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", r.Status, statusText(r.Status))

	// Headers
	if _, ok := r.Headers["Content-Length"]; !ok {
		r.SetHeader("Content-Length", strconv.Itoa(len(r.Body)))
	}
	if _, ok := r.Headers["Content-Type"]; !ok {
		r.SetHeader("Content-Type", "text/plain; charset=utf-8")
	}
	if _, ok := r.Headers["Server"]; !ok {
		r.SetHeader("Server", "httpserver/"+version)
	}

	// Write headers (sorted for consistency)
	keys := make([]string, 0, len(r.Headers))
	for k := range r.Headers {
		keys = append(keys, k)
	}
	// Simple insertion sort (small n)
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	for _, k := range keys {
		fmt.Fprintf(w, "%s: %s\r\n", k, r.Headers[k])
	}

	// Blank line separates headers from body
	if _, err := w.WriteString("\r\n"); err != nil {
		return err
	}

	// Body
	if len(r.Body) > 0 {
		if _, err := w.Write(r.Body); err != nil {
			return err
		}
	}

	return w.Flush()
}

// ensure strings package is used (it is — via canonicalHeader)
var _ = strings.TrimSpace