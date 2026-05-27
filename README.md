# maestro-consumer

This is a tiny backend that uses [`github.com/justinush/maestro`](../maestro) the same way a real third-party app would: **separate module**, **REST**, **Postgres**, and **public APIs only**.

Why it exists: examples inside the Maestro repo can accidentally hide sharp edges. This one is meant to smoke out stuff like awkward imports, missing APIs, doc mismatches, or “wow implementing `run.Store` is annoying”.

---

## Prerequisites

- Go 1.26+
- Docker (for Postgres)
- Maestro repo checked out next to this repo at `../maestro` (used by `go.mod` `replace`)

---

## Quick start

```bash
cd maestro-consumer
cp .env.example .env
docker compose up -d
go mod tidy
go run ./cmd/api
```

On startup it runs the SQL files in `./migrations`.

---

## Environment variables

- **`DATABASE_URL`**: Postgres connection string  
  Example: `postgres://maestro:maestro@localhost:5433/maestro_consumer?sslmode=disable`
- **`ADDR`**: HTTP bind address (default `:8080`)
- **`WORKFLOW_PATH`**: workflow file path (default `workflow/kyc.yaml`)

---

## API

| Method | Path | Purpose |
|---|---|---|
| POST | `/kyc/start` | Start a new run |
| GET | `/kyc/{runID}` | Get status |
| GET | `/kyc/{runID}/events` | Get run trace events |
| POST | `/kyc/{runID}/profile` | Submit profile input |
| POST | `/kyc/{runID}/document` | Submit document input |
| POST | `/kyc/{runID}/review` | Submit manual review input |

---

## Try it (happy path)

Start a run and grab the `runId`:

```bash
BASE=http://localhost:8080

START=$(curl -s -X POST "$BASE/kyc/start")
echo "$START"
RUN=$(echo "$START" | jq -r .runId)
echo "runId=$RUN"
```

Submit profile:

```bash
curl -s -X POST "$BASE/kyc/$RUN/profile" \
  -H 'Content-Type: application/json' \
  -d '{"fullName":"Ada Lovelace","email":"ada@example.com"}' | jq .
```

Submit a document (anything except `passport` auto-approves in this demo):

```bash
curl -s -X POST "$BASE/kyc/$RUN/document" \
  -H 'Content-Type: application/json' \
  -d '{"documentType":"id_card","documentRef":"doc-1"}' | jq .
```

Fetch status and events:

```bash
curl -s "$BASE/kyc/$RUN" | jq .
curl -s "$BASE/kyc/$RUN/events" | jq .
```

---

## Manual review path

If you submit `passport`, the workflow pauses on manual review. Try it:

```bash
curl -s -X POST "$BASE/kyc/$RUN/document" \
  -H 'Content-Type: application/json' \
  -d '{"documentType":"passport","documentRef":"doc-2"}' | jq .
```

Then approve:

```bash
curl -s -X POST "$BASE/kyc/$RUN/review" \
  -H 'Content-Type: application/json' \
  -d '{"approved":true}' | jq .
```

---

## What this app validates (before Maestro v0.1.0)

- Using Maestro from a totally separate module with:

  ```go
  replace github.com/justinush/maestro => ../maestro
  ```

- Implementing `run.Store` outside the Maestro repo (Postgres + JSONB `workflow_runs.state`)
- The normal embed flow: `pkg/maestro` + `Runtime.RestoreInstance`
- Persist/restore loop: `run.RecordFromInstance` + revisioned `Save`

---

## After tagging Maestro v0.1.0

Re-test without the local replace:

1. Remove the `replace` line from `go.mod`
2. Fetch the tag:

```bash
go get github.com/justinush/maestro@v0.1.0
```

