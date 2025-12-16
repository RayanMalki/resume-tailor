-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------------------
-- 1) USERS
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL, -- bcrypt/argon2 hash
  display_name  TEXT,
  is_verified   BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- ----------------------------------------
-- 2) SESSIONS (for login)
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE, -- store hash of session token (never raw token)
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- ----------------------------------------
-- 3) PASSWORD RESET TOKENS
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS password_resets (
  id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE, -- hash only
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);

-- ----------------------------------------
-- 4) RESUMES (uploaded PDF + extracted text)
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS resumes (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  original_name TEXT NOT NULL,
  pdf_path      TEXT NOT NULL,
  extracted_text TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id);
CREATE INDEX IF NOT EXISTS idx_resumes_created_at ON resumes(created_at);

-- ----------------------------------------
-- 5) RUNS (one tailoring run = resume + job text)
-- ----------------------------------------
DO $$ BEGIN
  CREATE TYPE run_status AS ENUM ('created','queued','processing','failed','completed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS runs (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  resume_id     UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
  job_text      TEXT NOT NULL,
  status        run_status NOT NULL DEFAULT 'created',
  error_message TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_runs_user_id ON runs(user_id);
CREATE INDEX IF NOT EXISTS idx_runs_resume_id ON runs(resume_id);
CREATE INDEX IF NOT EXISTS idx_runs_status_created ON runs(status, created_at);

-- ----------------------------------------
-- 6) RUN_REPORTS (ATSReport + ChangePlan)
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS run_reports (
  run_id      UUID PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
  ats_report  JSONB NOT NULL,
  change_plan JSONB NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ----------------------------------------
-- 7) RUN_ARTIFACTS (ResumeSpec + LaTeX + PDF)
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS run_artifacts (
  run_id      UUID PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
  resume_spec JSONB NOT NULL,
  latex_path  TEXT NOT NULL,
  pdf_path    TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ----------------------------------------
-- 8) JOB QUEUE (DB-based background jobs)
-- ----------------------------------------
DO $$ BEGIN
  CREATE TYPE job_type AS ENUM ('process_run');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE job_status AS ENUM ('queued','running','failed','done');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS jobs (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  type         job_type NOT NULL,
  run_id       UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  status       job_status NOT NULL DEFAULT 'queued',
  attempts     INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  locked_by    TEXT,
  locked_at    TIMESTAMPTZ,
  last_error   TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_run_id ON jobs(run_id);

-- ----------------------------------------
-- 9) SUBSCRIPTIONS (for $/month)
-- Keep it generic; you can integrate Stripe later.
-- ----------------------------------------
DO $$ BEGIN
  CREATE TYPE subscription_status AS ENUM ('active','past_due','canceled','trialing');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS subscriptions (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status            subscription_status NOT NULL DEFAULT 'active',
  plan_code         TEXT NOT NULL, -- e.g. "basic_monthly"
  current_period_start TIMESTAMPTZ NOT NULL DEFAULT now(),
  current_period_end   TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '1 month'),
  cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,

  -- Stripe fields (optional until you add billing)
  stripe_customer_id TEXT,
  stripe_subscription_id TEXT,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

-- ----------------------------------------
-- 10) OPTIONAL: SIMPLE USAGE TRACKING (for limits)
-- ----------------------------------------
CREATE TABLE IF NOT EXISTS usage_events (
  id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  run_id     UUID REFERENCES runs(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL, -- "run_created", "pdf_generated", etc.
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_events_user_id_created ON usage_events(user_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- (optional) DROP statements go here later
-- +goose StatementEnd
