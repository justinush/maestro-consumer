CREATE TABLE IF NOT EXISTS vendor_sessions (
    external_ref    TEXT PRIMARY KEY,
    run_id          TEXT NOT NULL,
    expected_step   TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vendor_sessions_run_id ON vendor_sessions (run_id);

CREATE TABLE IF NOT EXISTS webhook_events (
    event_id    TEXT PRIMARY KEY,
    run_id      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
