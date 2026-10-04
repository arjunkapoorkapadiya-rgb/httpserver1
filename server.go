package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// globalRouter holds all routes. Set up in main() before starting.
var globalRouter *Router

func startServer(port int) error {
	addr := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer listener.Close()

	fmt.Printf("httpserver v%s listening on http://localhost:%d\n", version, port)
	fmt.Println("press Ctrl+C to stop")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	req, err := parseRequest(reader)
	if err != nil {
		fmt.Printf("[%s] parse error: %v\n", remote, err)
		resp := NewResponse(400)
		resp.Write("Bad Request\n")
		_ = resp.write(writer)
		return
	}

	// Log the request
	fmt.Printf("[%s] %s %s %s\n", remote, req.Method, req.Path, req.Version)

	var resp *Response
	if globalRouter != nil {
		resp = globalRouter.Handle(req)
	} else {
		resp = NewResponse(500)
		resp.Write("Router not initialized\n")
	}

	if err := resp.write(writer); err != nil {
		fmt.Printf("[%s] write error: %v\n", remote, err)
	}
}

// ==== Utility helpers for handlers ====

// writeJSON writes a JSON response.
func writeJSON(resp *Response, status int, data interface{}) {
	resp.Status = status
	resp.SetHeader("Content-Type", "application/json")
	b, err := json.Marshal(data)
	if err != nil {
		resp.Status = 500
		resp.WriteBytes([]byte(`{"error":"json marshal failed"}`))
		return
	}
	resp.WriteBytes(b)
}

// readJSON parses a JSON request body into v.
func readJSON(req *Request, v interface{}) error {
	if len(req.Body) == 0 {
		return fmt.Errorf("empty body")
	}
	return json.Unmarshal(req.Body, v)
}