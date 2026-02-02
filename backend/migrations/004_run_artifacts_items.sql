-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS run_artifacts_items (
  run_id    UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  type      TEXT NOT NULL,
  content   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (run_id, type)
);

CREATE INDEX IF NOT EXISTS idx_run_artifacts_items_run_id
  ON run_artifacts_items(run_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS run_artifacts_items;

-- +goose StatementEnd
