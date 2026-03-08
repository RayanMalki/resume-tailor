-- +goose Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS onboarding_seen BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE runs
  ADD COLUMN IF NOT EXISTS creator_ip TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_seen;
ALTER TABLE runs DROP COLUMN IF EXISTS creator_ip;
