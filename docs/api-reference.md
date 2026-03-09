# API Reference

## Base URL

```
/v1
```

In production: `https://www.resumetailor.live/api/v1`
In local dev: `http://localhost:8080/v1`

The Next.js frontend proxies `/api/*` requests to the Go API, so the browser always calls `/api/v1/...`.

---

## Authentication

All authenticated endpoints require an HttpOnly session cookie set by the login/signup endpoints. The cookie is sent automatically by the browser when using `credentials: "include"` (or equivalent).

**CSRF protection:** All mutating requests (`POST`, `PUT`, `DELETE`) require the header:
```
X-CSRF-Protection: 1
```

Requests missing this header are rejected with `403 Forbidden`.

---

## Error Format

All errors return a JSON body:
```json
{
  "error": "human-readable error message"
}
```

Common status codes:

| Code | Meaning |
|------|---------|
| `400` | Bad request / validation failure |
| `401` | Not authenticated |
| `403` | Forbidden (CSRF check failed or insufficient permission) |
| `404` | Resource not found |
| `409` | Conflict (e.g. email already in use) |
| `429` | Rate limit exceeded |
| `500` | Internal server error |

---

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| Global (all endpoints) | 60 req/min per IP |
| `POST /auth/login` | 5/min per IP |
| `POST /auth/signup` | 3/hour per IP |
| `POST /resumes/upload` | 10/min per IP |
| `POST /runs` (burst) | 1/min per user |
| `POST /runs` (daily) | 10/day per user |
| `POST /runs` (free-tier, no BYOK) | 3/day per user, 6/day per creator IP |

---

## Endpoints

### Utility

#### `GET /v1/`
Health/root response.

#### `GET /v1/health`
Returns `200 OK` when the service is up.

---

### Auth

#### `POST /v1/auth/signup`
Creates a new account. Sends a verification email if email provider is configured.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:** `201 Created`
```json
{
  "user_id": "uuid",
  "email": "user@example.com"
}
```

---

#### `POST /v1/auth/login`
Authenticates with email/password. Sets an HttpOnly session cookie.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:** `200 OK` + `Set-Cookie: session=...`

---

#### `POST /v1/auth/logout`
Invalidates the current session and clears the cookie.

**Auth required:** Yes

---

#### `GET /v1/auth/google/start`
Redirects to Google's OAuth consent screen.

---

#### `GET /v1/auth/google/callback`
Handles the OAuth callback from Google. Creates or finds the user account, sets session cookie, then redirects to the frontend.

---

#### `GET /v1/auth/verify?token=<token>`
Verifies a user's email address using the token sent in the verification email.

---

#### `POST /v1/auth/forgot-password`
Sends a password reset email.

**Request:**
```json
{ "email": "user@example.com" }
```

---

#### `POST /v1/auth/reset-password`
Resets the password using the token from the reset email.

**Request:**
```json
{
  "token": "reset-token",
  "password": "newpassword"
}
```

---

#### `POST /v1/auth/resend-verification`
Resends the email verification link.

**Auth required:** Yes

---

### User

#### `GET /v1/me`
Returns the current user's profile and summary stats.

**Auth required:** Yes

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "email_verified": true,
  "has_api_key": false,
  "onboarding_seen": true,
  "total_runs": 5,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

#### `PUT /v1/me/password`
Updates the current user's password.

**Auth required:** Yes

**Request:**
```json
{
  "current_password": "oldpass",
  "new_password": "newpass"
}
```

---

#### `DELETE /v1/me`
Permanently deletes the account and all associated data.

**Auth required:** Yes

---

#### `PUT /v1/me/api-key`
Saves a personal OpenAI API key (BYOK). The key is encrypted with AES-256-GCM before storage.

**Auth required:** Yes

**Request:**
```json
{ "api_key": "sk-..." }
```

---

#### `DELETE /v1/me/api-key`
Removes the stored OpenAI API key.

**Auth required:** Yes

---

#### `POST /v1/me/onboarding-seen`
Marks the onboarding flow as completed for the current user.

**Auth required:** Yes

---

### Resumes

#### `POST /v1/resumes`
Creates a resume from plain text.

**Auth required:** Yes

**Request:**
```json
{
  "text": "John Doe\nSoftware Engineer\n...",
  "label": "My Resume 2024"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "label": "My Resume 2024",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

#### `POST /v1/resumes/upload`
Uploads a `.pdf` or `.docx` resume file. Text is extracted server-side.

**Auth required:** Yes

**Request:** `multipart/form-data`
- `file` — the resume file (max 2 MB)
- `label` (optional) — display name

**Response:** `201 Created` — same shape as `POST /v1/resumes`

---

#### `GET /v1/resumes`
Lists all resumes for the current user.

**Auth required:** Yes

**Response:**
```json
[
  { "id": "uuid", "label": "My Resume", "created_at": "..." },
  ...
]
```

---

#### `GET /v1/resumes/{resumeID}`
Returns a single resume including its text content.

**Auth required:** Yes

**Response:**
```json
{
  "id": "uuid",
  "label": "My Resume",
  "text": "...",
  "created_at": "..."
}
```

---

### Runs

#### `POST /v1/runs`
Creates a new run (resume × job description analysis). Returns immediately; processing is async.

**Auth required:** Yes

**Request:**
```json
{
  "resume_id": "uuid",
  "job_description": "We are looking for a Senior Go Engineer...",
  "job_title": "Senior Go Engineer",
  "company_name": "Acme Corp",
  "discipline": "it_software",
  "project_controls": {
    "project-uuid-1": "pinned",
    "project-uuid-2": "exclude"
  }
}
```

Fields:
- `resume_id` — required; ID of a previously saved resume
- `job_description` — required; full job posting text
- `discipline` — optional; one of `it_software`, `mechanical`, `electrical`, `aerospace`, `industrial_logistics`
- `job_title`, `company_name` — optional; used in cover letter generation
- `project_controls` — optional; per-project `pinned | auto | exclude` controls

**Response:** `202 Accepted`
```json
{
  "run_id": "uuid",
  "status": "pending"
}
```

---

#### `GET /v1/runs`
Lists all runs for the current user (most recent first).

**Auth required:** Yes

**Response:**
```json
[
  {
    "id": "uuid",
    "status": "done",
    "discipline": "it_software",
    "score": 0.72,
    "created_at": "2024-01-01T00:00:00Z"
  },
  ...
]
```

---

#### `GET /v1/runs/{runID}`
Returns a single run including current status.

**Auth required:** Yes

**Response:**
```json
{
  "id": "uuid",
  "status": "done",
  "discipline": "it_software",
  "job_description": "...",
  "score": 0.72,
  "error_message": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:30Z"
}
```

`status` values: `pending` → `processing` → `done` | `error`

---

#### `GET /v1/runs/{runID}/report`
Returns the full BM25 + AI report for a completed run.

**Auth required:** Yes

**Response:**
```json
{
  "run_id": "uuid",
  "score": 0.72,
  "raw_coverage": 0.52,
  "discipline": "it_software",
  "profile_version": "v3",
  "bm25_signals": {
    "top_job_terms": [{ "term": "kubernetes", "score": 0.55, "category": "cloud_devops_db" }],
    "missing_job_terms": [{ "term": "terraform", "score": 0.48 }],
    "overlap_terms": ["docker", "golang"],
    "category_coverage": { "cloud_devops_db": 0.60, "languages": 0.85 },
    "bucketed_top_terms": { "cloud_devops_db": [...] }
  },
  "ats_summary": "Your resume shows strong alignment with Go and Docker requirements...",
  "ats_notes": ["Consider adding Terraform experience", "..."],
  "interview_questions": ["Describe a time you optimized a Kubernetes deployment...", "..."],
  "created_at": "2024-01-01T00:00:30Z"
}
```

---

#### `POST /v1/disciplines/detect`
Detects the most likely discipline for a given resume.

**Auth required:** Yes

**Request:**
```json
{ "resume_id": "uuid" }
```

**Response:**
```json
{
  "discipline": "it_software",
  "confidence": 0.87
}
```

---

### Artifacts

All artifact endpoints require auth and return the artifact as a file download or plain text.

| Method | Path | Response type |
|--------|------|--------------|
| `GET` | `/v1/runs/{runID}/artifacts/resume-latex` | `text/plain` (`.tex` source) |
| `GET` | `/v1/runs/{runID}/artifacts/resume-pdf` | `application/pdf` |
| `GET` | `/v1/runs/{runID}/artifacts/resume-docx` | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| `GET` | `/v1/runs/{runID}/artifacts/cover-letter` | `text/plain` |
| `GET` | `/v1/runs/{runID}/artifacts/cover-letter-pdf` | `application/pdf` |
| `GET` | `/v1/runs/{runID}/artifacts/cover-letter-docx` | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| `GET` | `/v1/runs/{runID}/artifacts/project-reasons` | `application/json` |

PDF artifacts return `404` if `RESUME_PDF_ENABLED` is not set to `1` on the worker.

**Auth required:** Yes for all artifact endpoints.
