import Link from "next/link";
import EditorialNav from "./components/EditorialNav";

const workflow = [
  {
    title: "Bring your baseline CV",
    detail: "Keep one source profile and tailor it per role instead of rewriting from scratch.",
  },
  {
    title: "Paste a job post",
    detail: "We surface missing keywords, weak phrasing, and discipline-specific gaps.",
  },
  {
    title: "Ship with confidence",
    detail: "Preview and export PDF or DOCX with a clear ATS breakdown you can explain.",
  },
];


export default function HomePage() {
  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="public" active="home" />

        <main className="mt-4 space-y-4 sm:mt-6 sm:space-y-6">
          <section className="grid gap-4 lg:grid-cols-[1.25fr_1fr]">
            <article className="rt-panel rt-fade-up p-6 sm:p-8" style={{ animationDelay: "40ms" }}>
              <p className="rt-label">Human-first resume builder</p>
              <h1 className="mt-4 font-grotesk text-4xl font-semibold leading-[0.95] text-[var(--rt-ink-900)] sm:text-5xl lg:text-7xl">
                Resumes that sound like you,
                <br />
                not a chatbot.
              </h1>
              <p className="font-serif-display mt-5 max-w-2xl text-xl leading-relaxed text-[var(--rt-ink-700)] sm:text-2xl">
                Resume Tailor helps you move fast without flattening your voice.
                You keep final control line-by-line while the app handles ATS signal and structure.
              </p>

              <div className="mt-7 flex flex-wrap items-center gap-3">
                <Link
                  href="/dashboard"
                  className="rt-btn-primary px-6 py-3 text-xs font-semibold uppercase tracking-[0.24em]"
                >
                  Open dashboard
                </Link>
                <Link
                  href="/dashboard"
                  className="rt-btn-secondary px-6 py-3 text-xs font-semibold uppercase tracking-[0.24em]"
                >
                  Explore ATS report
                </Link>
              </div>

              <p className="rt-chip mt-5">No templates. No jargon stuffing.</p>
            </article>

            <aside className="rt-panel-muted rt-fade-up p-6 sm:p-7" style={{ animationDelay: "120ms" }}>
              <h2 className="font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)] sm:text-5xl lg:text-6xl">
                How a run works
              </h2>
              <div className="mt-5 space-y-3">
                {workflow.map((item, index) => (
                  <article
                    key={item.title}
                    className="rounded-[1.35rem] border border-[rgba(50,55,66,0.16)] bg-[rgba(255,255,255,0.57)] px-4 py-4 sm:px-5"
                  >
                    <div className="flex items-start gap-3">
                      <span
                        className="mt-0.5 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-[var(--rt-blue)] text-xs font-semibold text-white"
                        aria-hidden="true"
                      >
                        {index + 1}
                      </span>
                      <div>
                        <h3 className="font-grotesk text-2xl font-semibold text-[var(--rt-ink-900)]">
                          {item.title}
                        </h3>
                        <p className="mt-1 font-serif-display text-lg leading-relaxed text-[var(--rt-ink-700)]">
                          {item.detail}
                        </p>
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            </aside>
          </section>

          <section className="grid gap-4 md:grid-cols-3">
            <article className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "180ms" }}>
              <p className="rt-label">Average first draft</p>
              <p className="mt-3 font-serif-display text-3xl text-[var(--rt-ink-700)]">
                <span className="text-4xl font-semibold text-[var(--rt-ink-900)]">under 3 min</span> from paste to final-ready version.
              </p>
            </article>
            <article className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "220ms" }}>
              <p className="rt-label">Editing confidence</p>
              <p className="mt-3 font-serif-display text-3xl text-[var(--rt-ink-700)]">
                <span className="text-4xl font-semibold text-[var(--rt-ink-900)]">89%</span> keep over half of their original wording.
              </p>
            </article>
            <article className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "260ms" }}>
              <p className="rt-label">Tone check</p>
              <p className="mt-3 font-serif-display text-3xl text-[var(--rt-ink-700)]">
                <span className="text-4xl font-semibold text-[var(--rt-ink-900)]">Balanced</span> language that avoids robotic keyword spam.
              </p>
            </article>
          </section>

        </main>
      </div>
    </div>
  );
}
