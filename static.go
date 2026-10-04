package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// mimeTypes maps file extensions to content types.
var mimeTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".htm":  "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript; charset=utf-8",
	".json": "application/json",
	".txt":  "text/plain; charset=utf-8",
	".xml":  "application/xml",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".ico":  "image/x-icon",
	".webp": "image/webp",
	".pdf":  "application/pdf",
	".woff": "font/woff",
	".woff2": "font/woff2",
	".ttf":  "font/ttf",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".wasm": "application/wasm",
	".zip":  "application/zip",
	".gz":   "application/gzip",
}

// contentTypeFor returns the MIME type for a file path.
func contentTypeFor(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ct, ok := mimeTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// serveStatic serves a file from a root directory.
// The requested path is relative to `rootDir`.
// Protects against path traversal ("..").
func serveStatic(rootDir, urlPath string) *Response {
	// Clean the URL path and prevent traversal
	cleanPath := filepath.Clean("/" + urlPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// Build the full filesystem path
	fullPath := filepath.Join(rootDir, cleanPath)

	// Security: ensure the resolved path is still inside rootDir
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return serverError("invalid root")
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return serverError("invalid path")
	}
	if !strings.HasPrefix(absFull, absRoot+string(filepath.Separator)) &&
		absFull != absRoot {
		return notFound(urlPath)
	}

	// Stat the file
	info, err := os.Stat(fullPath)
	if err != nil {
		return notFound(urlPath)
	}

	// If directory, try index.html
	if info.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		if idxInfo, err := os.Stat(indexPath); err == nil && !idxInfo.IsDir() {
			fullPath = indexPath
			info = idxInfo
		} else {
			return forbidden("directory listing not allowed")
		}
	}

	// Read the file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return serverError("failed to read file")
	}

	resp := NewResponse(200)
	resp.SetHeader("Content-Type", contentTypeFor(fullPath))
	resp.SetHeader("Cache-Control", "public, max-age=3600")
	resp.WriteBytes(data)
	return resp
}

// Helper responses

func notFound(path string) *Response {
	resp := NewResponse(404)
	resp.Write("Not Found: " + path + "\n")
	return resp
}

func forbidden(msg string) *Response {
	resp := NewResponse(403)
	resp.Write(msg + "\n")
	return resp
}

func serverError(msg string) *Response {
	resp := NewResponse(500)
	resp.Write(msg + "\n")
	return resp
}

// staticHandler builds a handler that serves files from rootDir.
// URL path after the prefix is used as the file path.
// Example: StaticHandler("./public") mounted at /static
//   GET /static/style.css → ./public/style.css
func staticHandler(rootDir string) Handler {
	return func(req *Request) *Response {
		// Wildcard param "path" contains everything after the mount point
		filePath := req.Query["path"]
		if filePath == "" {
			filePath = "index.html"
		}
		return serveStatic(rootDir, filePath)
	}
}

// ensure fmt is used
var _ = fmt.Sprintf