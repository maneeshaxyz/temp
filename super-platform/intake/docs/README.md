# Intake

Minimal Go HTTP service for receiving silver instance events.

## What it does

- Exposes `POST /v1/silver/events`
- Requires `Content-Type: application/json`
- Requires these fields:
  - `timestamp` (RFC3339 string)
  - `instance_id` (non-empty string)
  - `signature_version` (non-empty string)
  - `signature_updated_at` (RFC3339 string)
- Returns `202 Accepted` when valid

No auth is implemented yet.

## Run instructions

Prerequisite: Go `1.26+`.

Run from repository root:

```bash
go run ./super-platform/intake/cmd/intake
```

Run from the `intake` directory:

```bash
cd super-platform/intake
go run ./cmd/intake
```

Run on a custom port:

```bash
PORT=9090 go run ./super-platform/intake/cmd/intake
```

Build and run a binary:

```bash
cd super-platform/intake
go build -o bin/intake ./cmd/intake
./bin/intake
```

## Example request

```bash
curl -i \
  -X POST http://localhost:8080/v1/silver/events \
  -H 'Content-Type: application/json' \
  -d '{"timestamp":"2026-03-09T03:58:07Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}'
```

## Response behavior

- `202` valid payload with body `{"ok":true}`
- `400` malformed JSON or missing/invalid required fields
- `405` method not allowed
- `415` unsupported media type

## Push to GHCR

```bash
cd super-platform/intake
IMAGE=ghcr.io/<owner>/silver-intake

echo <github_pat_with_write_packages> | docker login ghcr.io -u <github_username> --password-stdin
docker build -t $IMAGE:latest .
docker push $IMAGE:latest
```
