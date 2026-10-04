package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Request represents a parsed HTTP request.
type Request struct {
	Method  string
	Path    string
	Query   map[string]string
	Version string
	Headers map[string]string
	Body    []byte
}

// parseRequest reads one HTTP request from the reader.
func parseRequest(reader *bufio.Reader) (*Request, error) {
	req := &Request{
		Query:   map[string]string{},
		Headers: map[string]string{},
	}

	// 1. Read request line: "GET /path HTTP/1.1\r\n"
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read request line: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return nil, fmt.Errorf("empty request line")
	}

	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line: %q", line)
	}
	req.Method = parts[0]
	req.Version = parts[2]

	// Split path and query string
	rawPath := parts[1]
	if qIdx := strings.IndexByte(rawPath, '?'); qIdx >= 0 {
		req.Path = rawPath[:qIdx]
		req.Query = parseQuery(rawPath[qIdx+1:])
	} else {
		req.Path = rawPath
	}

	// 2. Read headers until blank line
	for {
		hLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		hLine = strings.TrimRight(hLine, "\r\n")
		if hLine == "" {
			break
		}
		colon := strings.IndexByte(hLine, ':')
		if colon < 0 {
			continue // skip malformed headers
		}
		key := strings.TrimSpace(hLine[:colon])
		val := strings.TrimSpace(hLine[colon+1:])
		// Normalize header names to canonical form (Content-Type)
		req.Headers[canonicalHeader(key)] = val
	}

	// 3. Read body if Content-Length is set
	if clStr, ok := req.Headers["Content-Length"]; ok {
		cl, err := strconv.Atoi(clStr)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %q", clStr)
		}
		if cl > 0 {
			body := make([]byte, cl)
			if _, err := io.ReadFull(reader, body); err != nil {
				return nil, fmt.Errorf("read body: %w", err)
			}
			req.Body = body
		}
	}

	return req, nil
}

// parseQuery parses "a=1&b=2" into {"a": "1", "b": "2"}.
func parseQuery(s string) map[string]string {
	out := map[string]string{}
	if s == "" {
		return out
	}
	for _, pair := range strings.Split(s, "&") {
		if pair == "" {
			continue
		}
		if eq := strings.IndexByte(pair, '='); eq >= 0 {
			out[urlDecode(pair[:eq])] = urlDecode(pair[eq+1:])
		} else {
			out[urlDecode(pair)] = ""
		}
	}
	return out
}

// urlDecode handles %XX and + (space).
func urlDecode(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '+':
			sb.WriteByte(' ')
		case c == '%' && i+2 < len(s):
			hi := hexVal(s[i+1])
			lo := hexVal(s[i+2])
			if hi >= 0 && lo >= 0 {
				sb.WriteByte(byte(hi<<4 | lo))
				i += 2
			} else {
				sb.WriteByte(c)
			}
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// canonicalHeader converts "content-type" → "Content-Type".
func canonicalHeader(s string) string {
	parts := strings.Split(strings.ToLower(s), "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "-")
}