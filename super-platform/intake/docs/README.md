# Intake

`intake` is a small Go HTTP service that accepts events from silver instances and triggers a best-effort async callback back to that instance.

## Endpoint

`POST /v1/silver/events`

Required request headers:

- `Content-Type: application/json`

Required request fields:

- `timestamp`: RFC3339 string
- `instance_id`: non-empty string
- `signature_version`: non-empty string
- `signature_updated_at`: RFC3339 string

## Behavior

For a valid request, `intake`:

1. accepts the request
2. returns `202 Accepted` immediately
3. starts an async callback to `http://<instance_id>:8888/api/results`

Outbound callback headers:

- `Content-Type: application/json`
- `X-API-Key: <X_API_KEY>`

Outbound callback body:

```json
{
  "status": "success",
  "data": {
    "timestamp": "2026-03-05T10:30:45Z",
    "instance_id": "172.25.0.19",
    "signature_version": "daily.cld:0",
    "signature_updated_at": "2026-03-08T07:57:37Z"
  },
  "timestamp": "2026-03-05T10:30:45Z"
}
```

The async callback is best-effort only:

- failures are logged
- there are no retries
- there is no queue or worker pool
- process shutdown can interrupt in-flight callbacks

## Responses

- `202`: valid request, body `{"ok":true}`
- `400`: malformed JSON or missing required fields
- `405`: method not allowed
- `415`: unsupported media type

## Run

From the repository root:

```bash
go run ./super-platform/intake/cmd/intake
```

With a custom port:

```bash
PORT=9090 go run ./super-platform/intake/cmd/intake
```

With outbound callback auth:

```bash
X_API_KEY=<api-key> go run ./super-platform/intake/cmd/intake
```

From the `intake` directory:

```bash
cd super-platform/intake
go run ./cmd/intake
```

Build a binary:

```bash
cd super-platform/intake
go build -o bin/intake ./cmd/intake
./bin/intake
```

## Example

```bash
curl -i \
  -X POST http://localhost:8080/v1/silver/events \
  -H 'Content-Type: application/json' \
  -d '{"timestamp":"2026-03-09T03:58:07Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}'
```

## Container Push

```bash
cd super-platform/intake

echo <github_pat_with_write_packages> | docker login ghcr.io -u <github_username> --password-stdin

docker buildx build \
  --platform linux/amd64 \
  -t ghcr.io/silver-mail-platform/silver-platform-intake:latest \
  --push \
  .
```
