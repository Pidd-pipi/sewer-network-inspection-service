# Sewer Network Inspection Service

Provides sewer manhole inspection findings and a small follow-up workflow:
operations records for follow-up tasks, status transitions, stale-task
auto-closing, audit events, metrics, and a browser page.

## Layout

The service is split into configuration, domain types, in-memory storage,
validation, HTTP handlers, health handling, web serving, and the process
entrypoint under `backend/`. Data is intentionally local and resets on restart.

## Run and API

`PORT` selects the listen port and defaults to `8080`. From `backend/`, start with `go run .`.

- `GET /healthz` returns service health.
- `GET /api/inspections` returns current manhole findings.
- `POST /api/inspections/{id}/status` accepts `{"status":"open|scheduled|resolved"}`.
- `GET /ops/records`, `POST /ops/records`, `POST /ops/records/{id}/status`,
  `GET /ops/snapshot`, `GET /ops/rules` expose the follow-up workflow.
- `GET /` serves the browser page, which fetches the API.

## Verification

- `gofmt` applied.
- `go build ./...` passed.
- `go test ./...` passed, including collection, status change, invalid status, and missing-record HTTP paths.
- Runtime smoke: started with `PORT=18180 go run .`; health, collection, status update, ops routes, `/`, and `/app.js` each returned HTTP 200.
