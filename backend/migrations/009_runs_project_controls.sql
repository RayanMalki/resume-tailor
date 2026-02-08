-- +goose Up
-- +goose StatementBegin
ALTER TABLE runs
  ADD COLUMN IF NOT EXISTS project_controls JSONB NOT NULL DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE runs
  DROP COLUMN IF EXISTS project_controls;
-- +goose StatementEnd
