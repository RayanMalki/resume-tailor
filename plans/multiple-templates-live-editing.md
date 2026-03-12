# Plan: Multiple Resume Templates + Live Editing

## Context

The app currently generates a single-style resume (LaTeX "classic" layout), stores it as a PDF artifact, and shows it read-only in an iframe. Users have no way to adjust content after generation, and there's only one visual style. This plan adds:

1. Multiple resume templates (visual styles to pick from before or after generation)
2. Live resume editing (structured form to edit ResumeSpec fields, with instant HTML preview, and server re-render on save)

---

## Technology Decision: Keep LaTeX, Add HTML Preview Layer

LaTeX/Tectonic stays as the final PDF renderer — it produces the best typographic quality and is already working. However, LaTeX can't power live (instant) preview because Tectonic compilation takes 1–3 seconds per render.

**Solution:** Add React components as an HTML/CSS preview layer in the browser. Each template is:
- A React component (for instant browser preview during editing)
- A Go LaTeX renderer function (for final PDF download)

The preview doesn't need to be pixel-perfect — it just needs to faithfully represent the content. Users editing bullets care about content, not sub-millimeter margins.

**Rejected alternatives:**
- Puppeteer/WeasyPrint: adds a Node/Python process to prod, not worth replacing working LaTeX
- @react-pdf/renderer: custom layout engine, meaningful bundle size, preview diverges from PDF appearance

---

## Feature 1: Multiple Resume Templates

Templates to start with: `classic` (current), `modern`, `compact`

### Backend Changes

**`backend/internal/artifacts/types.go`**

Add per-template artifact constants:

```go
TypeResumeSpec         = "resume_spec"           // NEW: stores ResumeSpec JSON
TypeResumeLatexModern  = "resume_latex_modern"
TypeResumePDFModern    = "resume_pdf_modern"
TypeResumeLatexCompact = "resume_latex_compact"
TypeResumePDFCompact   = "resume_pdf_compact"
// classic keeps existing TypeResumeLatex / TypeResumePDF names
```

**`backend/internal/runs/types.go`**

Add `Template string` field to `Run` struct and `TemplateDefault = "classic"` constant. Add `ParseTemplate(raw string) (string, bool)` validator.

**`backend/internal/runs/repo.go`**

Include `template` column in CREATE and SELECT queries.

**Database migration:**

```sql
ALTER TABLE runs ADD COLUMN template TEXT NOT NULL DEFAULT 'classic';
```

**`backend/internal/latex/template.go`**

Change signature to `RenderResume(spec ai.ResumeSpec, template string) string`. Extract existing body into `renderClassic(spec, limits)`. Add `renderModern` and `renderCompact` (start as styled variants of classic — different fonts, margins, section header style). Switch dispatch at top of `RenderResume`.

**`backend/internal/jobs/worker.go`**
- After `GenerateResumeSpec` succeeds, marshal and store: `artifacts.InsertIfNotExists(ctx, runID, artifacts.TypeResumeSpec, specJSON)` — this is the canonical editable source
- Pass `runData.Template` to `latex.RenderResume(spec, runData.Template)`
- Use correct artifact type keys: `artifacts.ArtifactTypeForTemplate("latex", template)` helper

**`backend/internal/httpapi/handlers/runs_create.go`**

Add `Template string` to `CreateRunRequest`, validate against allowed values (`classic`, `modern`, `compact`), default to `"classic"`.

### New Endpoints

**`POST /v1/runs/{runID}/rerender`**

Body: `{ "template": "modern" }`
- Validates ownership
- Reads `resume_spec` artifact
- Calls `latex.RenderResume(spec, template)` → stores `resume_latex_{template}`
- Calls `latex.CompilePDF(latex)` → stores `resume_pdf_{template}` (synchronous, ~2s)
- Returns 200 with new artifact URL

**`GET /v1/runs/{runID}/artifacts/resume-spec`**

Returns the stored `ResumeSpec` JSON.

**`PATCH /v1/runs/{runID}/artifacts/resume-spec`**

Body: `{ "spec": ResumeSpec }`
- Upserts `resume_spec` artifact
- Deletes the current-template PDF artifact (invalidates cache)
- Synchronously re-renders LaTeX + PDF + DOCX for the run's current template
- Returns 200

Register all three in `backend/internal/httpapi/router.go`.

---

## Feature 2: Live Resume Editing

### Data Flow

```
AI generates ResumeSpec
  → stored as resume_spec artifact (JSON)
  → renders LaTeX, PDF, DOCX (initial generation, as today)

User opens "Edit" tab
  → frontend fetches GET /resume-spec
  → spec loaded into React state

User edits fields
  → React state updates instantly
  → ResumePreview re-renders in browser (no server call)

User clicks "Save"
  → PATCH /resume-spec with updated spec
  → server re-renders LaTeX + PDF + DOCX (~2s)
  → frontend shows toast, refreshes PDF preview
```

### Frontend Components

**`web/app/components/templates/TemplateClassic.tsx`**

HTML/CSS replica of the Classic LaTeX layout. Pure render component: `(spec: ResumeSpec) => JSX.Element`. Tailwind classes. A4 width fixed.

**`web/app/components/templates/TemplateModern.tsx`**

Two-column header variant.

**`web/app/components/templates/TemplateCompact.tsx`**

Tighter spacing variant.

**`web/app/components/ResumePreview.tsx`**

```tsx
export function ResumePreview({ spec, template }: { spec: ResumeSpec; template: string }) {
  // switch(template) → TemplateClassic | TemplateModern | TemplateCompact
  // Rendered in A4-sized div with CSS scale transform for screen fit
}
```

**`web/app/components/ResumeEditor.tsx`**

Client component. Accepts `initialSpec: ResumeSpec`, `runId: string`, `template: string`.
- `useState<ResumeSpec>` for local edit state
- Sections: Contact, Summary, Experience (per-bullet textareas), Skills, Education, Projects
- Each section has expand/collapse toggle
- Side-by-side layout: Editor left, ResumePreview right (live updates on every keystroke)
- "Save & Regenerate PDF" → PATCH `/resume-spec` → toast on success

**`web/app/result/[runId]/page.tsx`** (modify)
- Add `"edit"` tab to existing `"preview" | "report" | "cover" | "latex"` tab set
- Add template selector (radio group with thumbnails) above preview — calls `POST /rerender` on change, optimistically switches `ResumePreview` template instantly
- When `"edit"` tab selected: fetch `/resume-spec`, render `ResumeEditor` + `ResumePreview` side-by-side
- Extract `ResumeSpec` TypeScript type to `web/app/types/resume.ts` (currently only exists implicitly)

---

## Implementation Order

1. **Store ResumeSpec** — add `TypeResumeSpec` constant, store in worker after AI call, add `GET /resume-spec` handler
2. **DB migration** — `ALTER TABLE runs ADD COLUMN template TEXT NOT NULL DEFAULT 'classic'`
3. **Runs template field** — `types.go`, `repo.go`, `runs_create.go` handler
4. **Refactor latex renderer** — `RenderResume(spec, template)`, extract `renderClassic`, stub `renderModern`/`renderCompact`
5. **Worker plumbing** — pass template through, use template-specific artifact types
6. **`/rerender` endpoint + `PATCH /resume-spec` endpoint** — new handlers + router registration
7. **Frontend: ResumeSpec type + HTML template components** — `TemplateClassic`, `TemplateModern`, `TemplateCompact`, `ResumePreview`
8. **Frontend: ResumeEditor** — structured form with live preview
9. **Wire result page** — edit tab, template selector, save flow

---

## Critical Files

| File | Change |
|------|--------|
| `backend/internal/artifacts/types.go` | Add `TypeResumeSpec`, per-template artifact constants |
| `backend/internal/runs/types.go` | Add `Template` field, `ParseTemplate`, `TemplateDefault` |
| `backend/internal/runs/repo.go` | Include `template` in queries |
| `backend/internal/latex/template.go` | `RenderResume(spec, template)` + multi-template dispatch |
| `backend/internal/jobs/worker.go` | Store `resume_spec`, pass template, use template artifact types |
| `backend/internal/httpapi/handlers/runs_create.go` | Accept `template` field |
| `backend/internal/httpapi/router.go` | Register 3 new routes |
| `backend/web/app/result/[runId]/page.tsx` | Add edit tab, template selector |
| `backend/web/app/components/ResumePreview.tsx` | New: HTML live preview |
| `backend/web/app/components/ResumeEditor.tsx` | New: structured edit form |
| `backend/web/app/components/templates/*.tsx` | New: 3 template components |
| `backend/web/app/types/resume.ts` | New: shared `ResumeSpec` TS type |

---

## Verification

1. Submit a run → DB has `resume_spec` artifact with valid JSON
2. `GET /runs/{id}/artifacts/resume-spec` → returns correct spec
3. `POST /runs/{id}/rerender` with `template=modern` → new `resume_pdf_modern` artifact created, PDF downloads correctly
4. `PATCH /runs/{id}/artifacts/resume-spec` with edited spec → PDF regenerates with new content
5. Frontend: edit tab opens, form shows correct fields, live preview updates on keypress, Save completes in ~2-3s and new PDF loads
6. Template selector: switching template on result page updates preview and triggers re-render, PDF download reflects new template
