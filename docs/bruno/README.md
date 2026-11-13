# Server Monitor API Bruno collection

Open this directory as a collection in Bruno and select the `Local` environment.
It targets `http://localhost:8765`; update `baseUrl` in `environments/Local.bru`
if the API is listening elsewhere.

Start the API from the repository root:

```sh
go run .
```

The API listens on `localhost:8765` by default. Pass `-ip` and `-port` to
change the listening address or port.

- **List services** sends `GET /services`.
- **Filter services** sends `POST /services` with a JSON array of exact systemd
  unit names, for example `["ssh.service", "cron.service"]`. An empty array
  returns an empty list.
