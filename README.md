# maestro-consumer

This is a tiny backend that uses [`github.com/justinush/maestro`](../maestro) the same way a real third-party app would: **separate module**, **REST**, **Postgres**, and **public APIs only**.

Why it exists: examples inside the Maestro repo can accidentally hide sharp edges. This one is meant to smoke out stuff like awkward imports, missing APIs, doc mismatches, or persistence integration gaps before a Maestro release.

This demo validates **`pkg/workflow`** ([Maestro workflow registry](https://github.com/justinush/maestro/pull/13), on Maestro **v0.2.0+** / `main`). `go.mod` uses `replace` to `../maestro` until Maestro **v0.2.0** is tagged on the module path.

---

## Prerequisites

- Go 1.26+
- Docker (for Postgres)
- jq (optional, for nicer curl output)
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

On startup:

1. SQL in `./migrations` — application tables (`applicants`, …)
2. Maestro `postgres.ApplySchema` — `workflow_runs` for `run.Store` (demo convenience; production may copy [`schema.sql`](../maestro/pkg/run/postgres/schema.sql) into migrations instead)
3. `workflow.LoadDir` loads every `*.yaml` / `*.json` under **`WORKFLOWS_DIR`** (default `workflows/`)

---

## Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | `postgres://maestro:maestro@localhost:5433/maestro_consumer?sslmode=disable` | Postgres |
| `ADDR` | `:8080` | HTTP bind address |
| `WORKFLOWS_DIR` | `workflows` | Directory of workflow definition files |

---

## Workflows and routing

Workflow YAML lives under `workflows/`. Each file’s `id` + `version` is the registry key stored on runs.

Host routing (product keys → workflow) is in `internal/kyc/routes.go`:

| entity | flow | workflow id | Notes |
|--------|------|-------------|--------|
| SG | MAIN | `kyc.sg.main` | Full demo graph (profile → document → liveness → …) |
| SG | REFRESH | `kyc.sg.refresh` | Short assist flow (profile → done) |
| ID | MAIN | `kyc.id.main` | Same graph as SG main (spike) |

`POST /kyc/start` requires JSON `entity` and `flow` (case-insensitive). Unknown pairs return **400**.

---

## API

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/kyc/start` | Start a run (`entity`, `flow` in body) |
| GET | `/kyc/{runID}` | Get status |
| GET | `/kyc/{runID}/events` | Get run trace events |
| POST | `/kyc/{runID}/profile` | Submit profile input |
| POST | `/kyc/{runID}/document` | Submit document input (main flows only) |
| POST | `/kyc/{runID}/review` | Submit manual review input (main flows only) |

Status JSON includes `workflowId`, `workflowVersion`, and (on start) `entity` / `flow`.

---

## Try it (SG main — happy path)

```bash
BASE=http://localhost:8080

START=$(curl -s -X POST "$BASE/kyc/start" \
  -H 'Content-Type: application/json' \
  -d '{"entity":"SG","flow":"MAIN"}')
echo "$START" | jq .
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

Restart the API and `GET` the same `runId` again — restore goes through `workflow.Registry.RestoreInstance` using the persisted `workflowId` / `workflowVersion`.

---

## SG refresh (short flow)

```bash
curl -s -X POST "$BASE/kyc/start" \
  -H 'Content-Type: application/json' \
  -d '{"entity":"SG","flow":"REFRESH"}' | jq .

# Then POST /kyc/{runId}/profile only — no document/review steps on this graph.
```

---

## Manual review path (SG / ID main)

If you submit `passport`, the workflow pauses on manual review:

```bash
curl -s -X POST "$BASE/kyc/$RUN/document" \
  -H 'Content-Type: application/json' \
  -d '{"documentType":"passport","documentRef":"doc-2"}' | jq .

curl -s -X POST "$BASE/kyc/$RUN/review" \
  -H 'Content-Type: application/json' \
  -d '{"approved":true}' | jq .
```

---

## Persistence

| Data | Where |
|------|--------|
| Workflow runs (`run.Store`) | Maestro `pkg/run/postgres` → `workflow_runs` |
| Applicants (demo app data) | `migrations/` → `applicants` |

This demo calls `ApplySchema` on startup for simplicity:

```go
import "github.com/justinush/maestro/pkg/run/postgres"

postgres.ApplySchema(ctx, pool)
store := postgres.NewStore(pool)
```

In production, apply the same DDL via your migration tool (`postgres.SchemaDDL()` or a copied `schema.sql`), alongside your app tables.

---

## What this app validates

**Multi-workflow validation**

- `workflow.LoadDir` at startup — fail-fast if any file is invalid or keys collide
- Host route map → `workflow.Key` → `reg.NewInstance` / `reg.RestoreInstance`
- Dot workflow ids (`kyc.sg.main`, …) on `RunRecord` and API responses
- Single-workflow `maestro.Load` path unchanged in Maestro; this app uses the registry pattern only

**Embedding (ongoing)**

- Separate module with `replace github.com/justinush/maestro => ../maestro`
- Postgres `run.Store` + `RecordFromInstance` / revisioned `Save`
- This demo does **not** wrap workflow + app data in one DB transaction — production should

---

## After Maestro v0.2.0 is tagged

1. Remove the `replace` line from `go.mod`
2. Pin the module:

```bash
go get github.com/justinush/maestro@v0.2.0
go mod tidy
go run ./cmd/api
```

3. Re-run the curl flows above
