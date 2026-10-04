# httpserver

An HTTP/1.1 server built from scratch in Go. No `net/http` package. Just raw TCP sockets + our own protocol implementation.

## Features

- **HTTP/1.1 protocol** — request parser, response writer
- **Router** — pattern matching with path params (`/users/:id`) and wildcards (`/files/*path`)
- **Middleware system** — composable, chainable
  - `LoggingMiddleware` — logs each request with timing
  - `CORSMiddleware` — adds CORS headers
  - `AuthMiddleware` — Bearer token auth
  - `RecoveryMiddleware` — catches panics, returns 500
- **Static file serving** — with MIME types and path traversal protection
- **JSON helpers** — `writeJSON`, `readJSON`
- **Concurrency** — goroutine-per-connection
- **Query string parsing** — with URL decoding
- **Command-line interface** — port, static dir configurable

## Installation

```bash
git clone https://github.com/arjunkapoorkapadiya-rgb/httpserver.git
cd httpserver
go build -o httpserver.exe
```

## Usage

```bash
# Start on default port 8080, serve ./public as static files
httpserver start

# Custom port and static dir
httpserver start --port 3000 --static ./www
```

## Example

```bash
# Home
curl http://localhost:8080/
# → Welcome HTML page

# Query param
curl "http://localhost:8080/greet?name=Chinmay"
# → Hello, Chinmay!

# Path param
curl http://localhost:8080/users/42
# → User 42

# JSON API
curl http://localhost:8080/api/info
# → {"server":"httpserver","status":"running","version":"0.5.0"}

# POST echo
curl -X POST -d "hello" http://localhost:8080/echo
# → hello

# Static file
curl http://localhost:8080/static/index.html
# → HTML content

# Auth-protected route
curl http://localhost:8080/admin
# → Unauthorized
curl -H "Authorization: Bearer secret123" http://localhost:8080/admin
# → Welcome to the admin panel!
```

## Architecture

```
       TCP socket
           │
           ▼
    ┌─────────────┐
    │  accept()   │  goroutine per connection
    └─────────────┘
           │
           ▼
    ┌─────────────┐
    │  Request    │  parse HTTP request
    │  Parser     │
    └─────────────┘
           │
           ▼
    ┌─────────────┐
    │  Middleware │  recovery → logging → cors → auth
    │  Chain      │
    └─────────────┘
           │
           ▼
    ┌─────────────┐
    │  Router     │  match method + path (+ params)
    └─────────────┘
           │
           ▼
    ┌─────────────┐
    │  Handler    │  returns *Response
    └─────────────┘
           │
           ▼
    ┌─────────────┐
    │  Response   │  write status + headers + body
    │  Writer     │
    └─────────────┘
```

## File map

| File | Purpose |
|---|---|
| `main.go` | CLI entry point, route setup |
| `server.go` | TCP listener, connection handling, JSON helpers |
| `request.go` | HTTP request parser |
| `response.go` | HTTP response writer |
| `router.go` | Router with pattern matching |
| `middleware.go` | Middleware system |
| `static.go` | Static file serving |
| `*_test.go` | Tests for parser, router, middleware |

## Testing

```bash
go test -v
```

Tests cover: request parsing, query strings, POST bodies, URL decoding, route matching, path params, wildcards, method mismatch, middleware ordering, auth, panic recovery.

## Why build this?

Building an HTTP server from scratch teaches:
- **TCP sockets** — accept, read, write
- **Protocol parsing** — HTTP/1.1 request format
- **Concurrency** — goroutine-per-connection
- **Middleware patterns** — chain of responsibility
- **Security** — path traversal, auth, panic recovery

## LOC

~1400 lines of Go across 6 source files and 3 test files.

## License

MIT