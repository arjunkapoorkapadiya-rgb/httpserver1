package main

import (
	"fmt"
	"os"
	"strconv"
)

const usage = `httpserver — an HTTP server built from scratch

usage:
  httpserver start [--port <port>] [--static <dir>]     start the server
  httpserver version                                    print version
  httpserver help                                       show this help

options:
  --port <n>       port to listen on (default: 8080)
  --static <dir>   directory to serve at /static/ (default: ./public)

examples:
  httpserver start
  httpserver start --port 3000
  httpserver start --static ./www
`

const version = "0.5.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "start":
		if err := cmdStart(args); err != nil {
			fatal(err)
		}
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fatal(fmt.Errorf("unknown command: %s", cmd))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "httpserver: %v\n", err)
	os.Exit(1)
}

func cmdStart(args []string) error {
	port := 8080
	staticDir := "./public"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("--port requires a value")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil || p < 1 || p > 65535 {
				return fmt.Errorf("invalid port: %s", args[i+1])
			}
			port = p
			i++
		case "--static", "-s":
			if i+1 >= len(args) {
				return fmt.Errorf("--static requires a value")
			}
			staticDir = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	setupRoutes(staticDir)
	return startServer(port)
}

func setupRoutes(staticDir string) {
	globalRouter = NewRouter()

	// Global middlewares
	globalRouter.Use(RecoveryMiddleware())
	globalRouter.Use(LoggingMiddleware())
	globalRouter.Use(CORSMiddleware())

	// === Public API routes ===

	globalRouter.GET("/", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.SetHeader("Content-Type", "text/html; charset=utf-8")
		resp.Write(`<!DOCTYPE html>
<html>
<head><title>httpserver</title></head>
<body>
<h1>Welcome to httpserver v` + version + `</h1>
<p>Try these routes:</p>
<ul>
  <li><a href="/hello">/hello</a></li>
  <li><a href="/greet?name=Chinmay">/greet?name=Chinmay</a></li>
  <li><a href="/users/42">/users/42</a></li>
  <li><a href="/api/info">/api/info</a></li>
  <li><a href="/static/">/static/</a> (static files)</li>
</ul>
</body>
</html>`)
		return resp
	})

	globalRouter.GET("/hello", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write("Hello, World!\n")
		return resp
	})

	globalRouter.GET("/greet", func(req *Request) *Response {
		name := req.Query["name"]
		if name == "" {
			name = "stranger"
		}
		resp := NewResponse(200)
		resp.Write("Hello, " + name + "!\n")
		return resp
	})

	globalRouter.GET("/users/:id", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write("User " + req.Query["id"] + "\n")
		return resp
	})

	globalRouter.GET("/users/:id/posts", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write("Posts for user " + req.Query["id"] + "\n")
		return resp
	})

	globalRouter.POST("/echo", func(req *Request) *Response {
		resp := NewResponse(200)
		resp.WriteBytes(req.Body)
		return resp
	})

	globalRouter.GET("/api/info", func(req *Request) *Response {
		resp := NewResponse(200)
		writeJSON(resp, 200, map[string]interface{}{
			"server":  "httpserver",
			"version": version,
			"status":  "running",
			"static":  staticDir,
		})
		return resp
	})

	globalRouter.GET("/panic", func(req *Request) *Response {
		panic("intentional panic to test recovery middleware")
	})

	// Auth-protected route
	auth := AuthMiddleware("secret123")
	globalRouter.GET("/admin", auth(func(req *Request) *Response {
		resp := NewResponse(200)
		resp.Write("Welcome to the admin panel!\n")
		return resp
	}))

	// === Static file serving ===
	// GET /static/*path → serveStatic(staticDir, path)
	globalRouter.GET("/static/*path", staticHandler(staticDir))
}