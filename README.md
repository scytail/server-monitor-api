# Server Monitor API

Server Monitor API is a small Go service that exposes the host's systemd service state over a simple REST API. It can list all active and inactive `.service` units and filter the results to a subset of unit names.

This project is useful for monitoring and querying local system services from another tool or script without directly interacting with systemd.

## Requirements

- Go 1.22 or newer
- Linux with systemd installed and running
- Access to the systemd D-Bus interface for the local machine

The service connects to systemd using the `go-systemd` DBus client, so it is intended for a Linux environment with systemd available.

## Build

From the repository root:

```bash
go mod download
go build ./...
```

## Run

Start the API with the default host and port:

```bash
go run .
```

The server listens on `localhost:8765` by default.

You can override the bind address and port:

```bash
go run . -ip 0.0.0.0 -port 8080
```

## Test

Run the test suite:

```bash
go test ./...
```

## API

### List all services

```bash
curl http://localhost:8765/services
```

### Filter services by name

```bash
curl -X POST http://localhost:8765/services \
  -H 'Content-Type: application/json' \
  -d '["ssh.service", "cron.service"]'
```

The request body must be a JSON array of exact systemd unit names. An empty array returns an empty list.

## Project layout

- `main.go` starts the HTTP server
- `internal/api` contains the HTTP handlers and JSON response logic
- `internal/services` contains the systemd integration and filtering logic
- `docs/bruno` contains Bruno API request examples
