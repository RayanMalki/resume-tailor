-- +goose Up
-- +goose StatementBegin

ALTER TABLE resumes
  ADD COLUMN IF NOT EXISTS title TEXT,
  ADD COLUMN IF NOT EXISTS content_text TEXT,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE resumes
SET title = COALESCE(title, original_name, ''),
    content_text = COALESCE(content_text, extracted_text, '');

ALTER TABLE resumes
  ALTER COLUMN title SET NOT NULL,
  ALTER COLUMN content_text SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE resumes
  ALTER COLUMN title DROP NOT NULL,
  ALTER COLUMN content_text DROP NOT NULL;

-- +goose StatementEnd
