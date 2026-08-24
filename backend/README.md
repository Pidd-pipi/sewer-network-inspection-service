# Sewer Network Inspection Backend

This directory contains the Go module and service entrypoint.

```bash
go test ./...
go build ./...
go run .
```

The `main` package starts the HTTP service on `PORT` (default `8080`).

## Endpoints

- `GET /healthz` — service health
- `GET /api/inspections` — current manhole findings
- `POST /api/inspections/{id}/status` — update a finding status (`open|scheduled|resolved`)
- `GET /ops/records` — list operations records with optional `subject`/`status`/`priority`/`owner`/`page`/`pageSize`
- `POST /ops/records` — create an operations record (requires `owner`, `priority`, and a `site` label)
- `POST /ops/records/{id}/status` — transition an operations record (`queued|active|paused|closed`, optional `expectedRevision`)
- `GET /ops/snapshot` — record counts by status and priority
- `GET /ops/rules` — the inspection control-rule list
- `GET /` and `GET /app.js` — browser page

The operations workflow is layered into domain model, validation, status
transition, concurrency-safe storage, audit events, batch/worker automation,
metrics, and HTTP lifecycle handling. Requests carry a request id and pass
through recovery, timeout, and metrics middlewares.
