CREATE TABLE IF NOT EXISTS applicants (
    applicant_id TEXT PRIMARY KEY,
    run_id       TEXT NOT NULL UNIQUE,
    profile      JSONB,
    documents    JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_applicants_run_id ON applicants (run_id);

CREATE TABLE IF NOT EXISTS workflow_runs (
    run_id           TEXT PRIMARY KEY,
    workflow_id      TEXT NOT NULL,
    workflow_version TEXT NOT NULL,
    revision         BIGINT NOT NULL,
    state            JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_revision ON workflow_runs (run_id, revision);
