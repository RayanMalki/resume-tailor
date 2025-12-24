-- +goose Up
-- +goose StatementBegin

-- Make current resumes table compatible with "text resume" MVP (title + content_text)

ALTER TABLE resumes
  ADD COLUMN IF NOT EXISTS title TEXT,
  ADD COLUMN IF NOT EXISTS content_text TEXT,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- These were required for PDF upload flow; make them optional for text-only MVP
ALTER TABLE resumes
  ALTER COLUMN original_name DROP NOT NULL,
  ALTER COLUMN pdf_path DROP NOT NULL,
  ALTER COLUMN extracted_text DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_resumes_user_id_created_at
  ON resumes(user_id, created_at DESC);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- Revert text-resume MVP changes (best-effort)
DROP INDEX IF EXISTS idx_resumes_user_id_created_at;

ALTER TABLE resumes
  DROP COLUMN IF EXISTS title,
  DROP COLUMN IF EXISTS content_text,
  DROP COLUMN IF EXISTS updated_at;

-- NOTE: We don't re-add NOT NULL constraints automatically here
-- because old rows might violate them.

-- +goose StatementEnd
