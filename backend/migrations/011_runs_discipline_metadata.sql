-- +goose Up
-- +goose StatementBegin
ALTER TABLE runs
  ADD COLUMN IF NOT EXISTS discipline TEXT,
  ADD COLUMN IF NOT EXISTS discipline_confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS discipline_source TEXT NOT NULL DEFAULT 'auto';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_runs_discipline_source'
  ) THEN
    ALTER TABLE runs
      ADD CONSTRAINT chk_runs_discipline_source
      CHECK (discipline_source IN ('auto', 'user_override'));
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_runs_discipline'
  ) THEN
    ALTER TABLE runs
      ADD CONSTRAINT chk_runs_discipline
      CHECK (
        discipline IS NULL OR
        discipline IN ('mechanical', 'electrical', 'industrial_logistics', 'aerospace', 'it_software')
      );
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE runs
  DROP CONSTRAINT IF EXISTS chk_runs_discipline_source;
ALTER TABLE runs
  DROP CONSTRAINT IF EXISTS chk_runs_discipline;
ALTER TABLE runs
  DROP COLUMN IF EXISTS discipline_source,
  DROP COLUMN IF EXISTS discipline_confidence,
  DROP COLUMN IF EXISTS discipline;
-- +goose StatementEnd
