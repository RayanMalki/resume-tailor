-- +goose Up
-- +goose StatementBegin

DO $$ BEGIN
  ALTER TYPE run_status ADD VALUE IF NOT EXISTS 'running';
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  ALTER TYPE run_status ADD VALUE IF NOT EXISTS 'succeeded';
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

UPDATE runs SET status = 'queued' WHERE status = 'created';
UPDATE runs SET status = 'running' WHERE status = 'processing';
UPDATE runs SET status = 'succeeded' WHERE status = 'completed';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE runs SET status = 'completed' WHERE status = 'succeeded';
UPDATE runs SET status = 'processing' WHERE status = 'running';
UPDATE runs SET status = 'created' WHERE status = 'queued';

-- NOTE: We don't remove enum values in down migration.

-- +goose StatementEnd
