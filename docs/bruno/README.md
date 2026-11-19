# Server Monitor API Bruno collection

Open this directory as a collection in Bruno and select the `Local` environment.
It targets `http://localhost:8765`; update `baseUrl` in `environments/Local.bru`
if the API is listening elsewhere.

Start the API from the repository root:

```sh
CREDENTIALS_DIRECTORY=/path/to/systemd/credentials go run .
```

The API listens on `localhost:8765` by default. Pass `-ip` and `-port` to
change the listening address or port; do not bind to a non-loopback address
until HTTPS is configured. The credentials directory must contain
an `gomonitor-api-token` file with a 32-byte URL-safe base64 token (43
characters, without padding). Generate one with:

```sh
python3 -c 'import secrets; print(secrets.token_urlsafe(32))'
```

The blacklist is loaded at startup from
`/etc/server-monitor-api/blacklist`; use `-blacklist-file /path/to/blacklist`
to select another file. Put one exact service name or pattern per line; `*`
matches any number of characters. Blank lines and lines beginning with `#` are
ignored. A missing file is treated as an empty blacklist.

In the deployed systemd unit, use
`LoadCredential=gomonitor-api-token:/path/to/token-file`; systemd then sets
`CREDENTIALS_DIRECTORY` for the service. Keep the source credential file
readable only by the service administrator. The API refuses to start if the
credential is missing or invalid.

Set `apiToken` in the Bruno `Local` environment to the same token. The checked-in
value is only a placeholder; do not commit a real token.

- **List services** sends `GET /services`.
- **Filter services** sends `POST /services` with a JSON array of exact systemd
  unit names, for example `["ssh.service", "cron.service"]`. An empty array
  returns an empty list.

Both requests send `Authorization: Bearer <token>`. For example:

```sh
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8765/services
```
