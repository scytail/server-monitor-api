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

## Automated releases

Every successful push to `main` creates a GitHub Release tagged with the
GitHub Actions run number (for example, `build-42`). Each release contains
Linux amd64 and arm64 tarballs with the `server-monitor-api` executable, plus
a `SHA256SUMS` file for verifying the packages. Download the matching
architecture from the repository's Releases page before deploying.

## Run

Start the API with the default host and port. The API requires a token supplied
through a credential file named `gomonitor-api-token`; systemd exposes
credentials to the process through `CREDENTIALS_DIRECTORY`. 

### Local Testing

When testing locally, run the following command:

```bash
CREDENTIALS_DIRECTORY=/path/to/credentials/directory go run .
```

The server listens on `localhost:8765` by default.

You can override the bind address and port:

```bash
go run . -ip 127.0.0.1 -port 8080
```

The blacklist is loaded at startup from
`/etc/server-monitor-api/blacklist`. Override the location with
`-blacklist-file`. Put one exact service name or wildcard pattern per line;
`*` matches any number of characters, and blank lines and lines beginning
with `#` are ignored:

```bash
go run . -blacklist-file /path/to/blacklist
```

Blacklisted services are excluded from `GET /services` and removed before the
requested-name filtering performed by `POST /services`. If the file does not
exist, the API starts with an empty blacklist.

### Service Deployment

For systemd deployment, configure
`LoadCredential=gomonitor-api-token:/path/to/token-file/gomonitor-api-token`.
Save this line underneath the [Service] section of the systemd unit.
The file can be saved anywhere, but it is recommended to be saved in `etc/server-monitor-api/gomonitor-api-token`.
The credential must contain a 32-byte URL-safe base64 token (43 characters,
without padding). Generate one with:

```bash
python3 -c 'import secrets; print(secrets.token_urlsafe(32))' >> /path/to/token-file/gomonitor-api-token
```

Keep the source credential file private, and do not commit the token. The gitignore safely ignores testing files located in a `.test` folder, so for local development purposes, the token file can be stored in there. The
service exits at startup if the credential is missing or malformed. Since the
API currently listens on localhost by default and is intended for same-host
clients, HTTPS is not configured. Do not bind to a non-loopback address until
HTTPS is configured.

## Test

Run the test suite:

```bash
go test ./...
```

## API

### List all services

```bash
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8765/services
```

If using the included bruno collection, make sure to update the `apiToken` variable to include the token generated and stored in the relevant file.

### Filter services by name

```bash
curl -X POST http://localhost:8765/services \
  -H "Authorization: Bearer $API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '["ssh.service", "cron.service"]'
```

The request body must be a JSON array of exact systemd unit names. An empty array returns an empty list.

## Project layout

- `main.go` starts the HTTP server
- `api` contains the HTTP handlers and JSON response logic
- `internal/services` contains the systemd integration and filtering logic
- `docs/bruno` contains Bruno API request examples
