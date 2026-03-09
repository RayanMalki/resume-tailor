# Frontend (Next.js)

The web frontend is a Next.js App Router application (TypeScript + Tailwind CSS). It communicates with the Go API via a built-in `/api/*` reverse proxy.

---

## Tech Stack

| Technology | Role |
|-----------|------|
| Next.js 14+ (App Router) | Framework, routing, SSR |
| TypeScript | Type safety |
| Tailwind CSS | Styling |
| Next.js API routes | Proxy `/api/*` → Go API |

---

## Pages and Routes

| Route | Description |
|-------|-------------|
| `/` | Landing page |
| `/login` | Email/password and Google OAuth login |
| `/welcome` | Post-signup onboarding |
| `/dashboard` | Run history, ATS score trend chart |
| `/resume` | Upload or paste a resume |
| `/job` | Enter job description and discipline |
| `/result/[runId]` | Results page: ATS report, BM25 analysis, resume, cover letter |
| `/settings` | Account settings, BYOK key management |
| `/forgot-password` | Request a password reset email |
| `/reset-password` | Set a new password via email token |
| `/verify` | Email verification landing |

---

## Key Components

Components live in `app/components/`. Notable ones:

- **`EditorialNav`** — top navigation bar
- **`RunResultTabs`** (or equivalent) — multi-tab results view on `/result/[runId]`
- **`ATSScoreCard`** — displays the BM25 coverage score and breakdown
- **`ResumeUpload`** — drag-and-drop file upload component
- **`DisciplineSelector`** — dropdown for discipline selection

---

## Environment Variables

Create `web/.env.local` (not committed):

```bash
# Required
NEXT_PUBLIC_API_BASE_URL=/api        # Base path for all API calls (browser-side)
API_PROXY_TARGET=http://localhost:8080 # Where to proxy /api/* requests (server-side)

# Optional
NEXT_PUBLIC_SITE_URL=http://localhost:3000  # Used for canonical URLs / OpenGraph
```

In production (via `render.yaml`):
```bash
NEXT_PUBLIC_API_BASE_URL=/api
API_PROXY_TARGET=https://resume-tailor-api-m106.onrender.com
NEXT_PUBLIC_SITE_URL=https://www.resumetailor.live
```

---

## Running Locally

Prerequisites: Node.js 18+, backend API running at `localhost:8080`.

```bash
cd web
npm install
npm run dev
```

The app is available at `http://localhost:3000`.

All `/api/*` requests are proxied to `API_PROXY_TARGET` so the session cookie flows correctly (same-origin from the browser's perspective).

---

## Building for Production

```bash
npm run build
npm run start
```

Or let Render build it: `npm ci --include=dev && npm run build`.

---

## Notes

- All API calls use `credentials: "include"` so the session cookie is sent automatically.
- The proxy approach avoids cross-origin issues with cookies in production (no `SameSite=None` headaches in development).
- The Go API must have `FRONTEND_ORIGIN` set to the frontend origin to allow credentialed CORS requests.
