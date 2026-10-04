package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseRequestBasic(t *testing.T) {
	raw := "GET /hello HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"User-Agent: test/1.0\r\n" +
		"\r\n"

	reader := bufio.NewReader(strings.NewReader(raw))
	req, err := parseRequest(reader)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("method: got %q, want %q", req.Method, "GET")
	}
	if req.Path != "/hello" {
		t.Errorf("path: got %q, want %q", req.Path, "/hello")
	}
	if req.Version != "HTTP/1.1" {
		t.Errorf("version: got %q, want %q", req.Version, "HTTP/1.1")
	}
	if req.Headers["Host"] != "localhost:8080" {
		t.Errorf("Host header: got %q", req.Headers["Host"])
	}
	if req.Headers["User-Agent"] != "test/1.0" {
		t.Errorf("User-Agent header: got %q", req.Headers["User-Agent"])
	}
}

func TestParseRequestQuery(t *testing.T) {
	raw := "GET /search?q=hello&limit=10 HTTP/1.1\r\nHost: x\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	req, err := parseRequest(reader)
	if err != nil {
		t.Fatal(err)
	}
	if req.Path != "/search" {
		t.Errorf("path: got %q", req.Path)
	}
	if req.Query["q"] != "hello" {
		t.Errorf("q: got %q", req.Query["q"])
	}
	if req.Query["limit"] != "10" {
		t.Errorf("limit: got %q", req.Query["limit"])
	}
}

func TestParseRequestPOST(t *testing.T) {
	body := `{"name":"x"}`
	raw := "POST /api HTTP/1.1\r\n" +
		"Host: x\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		body

	reader := bufio.NewReader(strings.NewReader(raw))
	req, err := parseRequest(reader)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != "POST" {
		t.Errorf("method: got %q", req.Method)
	}
	if string(req.Body) != body {
		t.Errorf("body: got %q, want %q", string(req.Body), body)
	}
}

func TestParseRequestURLDecode(t *testing.T) {
	raw := "GET /search?q=hello%20world&name=Chinmay+Lahoti HTTP/1.1\r\nHost: x\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	req, _ := parseRequest(reader)
	if req.Query["q"] != "hello world" {
		t.Errorf("q: got %q, want %q", req.Query["q"], "hello world")
	}
	if req.Query["name"] != "Chinmay Lahoti" {
		t.Errorf("name: got %q, want %q", req.Query["name"], "Chinmay Lahoti")
	}
}

func TestParseRequestBadRequestLine(t *testing.T) {
	raw := "GARBAGE\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	_, err := parseRequest(reader)
	if err == nil {
		t.Error("expected error for malformed request line, got nil")
	}
}