-- +goose Up
-- +goose StatementBegin

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS auth_provider TEXT NOT NULL DEFAULT 'local',
  ADD COLUMN IF NOT EXISTS oauth_provider TEXT,
  ADD COLUMN IF NOT EXISTS oauth_sub TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_oauth_provider_sub
  ON users(oauth_provider, oauth_sub)
  WHERE oauth_provider IS NOT NULL AND oauth_sub IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- (optional) DROP statements go here later
-- +goose StatementEnd
