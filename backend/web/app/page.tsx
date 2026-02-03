import Link from "next/link";
import TypewriterHero from "./components/TypewriterHero";

export default function HomePage() {
  return (
    <div className="relative min-h-screen overflow-hidden bg-ink-950 text-slate-100">
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute left-[-20%] top-[-30%] h-[520px] w-[520px] rounded-full bg-glow-orange opacity-50 blur-[140px]" />
        <div className="absolute right-[-10%] top-10 h-[420px] w-[420px] rounded-full bg-glow-red opacity-40 blur-[140px]" />
        <div className="absolute bottom-[-20%] left-1/3 h-[520px] w-[520px] rounded-full bg-glow-amber opacity-40 blur-[160px]" />
        <div className="absolute inset-0 bg-grid opacity-30" />
      </div>

      <header className="relative mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-ember-500/60 bg-ink-900 text-lg font-semibold text-ember-300 shadow-glow">
            RT
          </span>
          <div>
            <div className="text-lg font-semibold text-white">Resume Tailor</div>
            <div className="text-xs uppercase tracking-[0.3em] text-ember-400/80">
              laser-fit resumes
            </div>
          </div>
        </div>
        <Link
          href="/login"
          className="rounded-full border border-ember-500/60 bg-ink-900/70 px-5 py-2 text-sm font-medium text-ember-200 shadow-glow transition hover:-translate-y-0.5 hover:border-ember-400 hover:text-white"
        >
          Login
        </Link>
      </header>

      <main className="relative mx-auto w-full max-w-6xl px-6 pb-20 pt-8">
        <section className="grid gap-10 lg:grid-cols-[1.1fr_0.9fr]">
          <div className="rounded-[32px] border border-white/10 bg-ink-900/70 p-10 shadow-panel backdrop-blur">
            <p className="text-sm font-semibold uppercase tracking-[0.4em] text-ember-300/80">
              Welcome
            </p>
            <h1 className="mt-4 text-3xl font-semibold text-white sm:text-4xl">
              Ship a tailored resume in minutes.
            </h1>
            <TypewriterHero
              className="mt-4 text-lg font-medium text-ember-200 sm:text-xl"
              phrases={[
                "Tailor your resume to every job in minutes.",
                "Beat ATS algorithms to boost your interview odds.",
                "Turn job posts into focused, ATS-ready bullets.",
                "Surface missing keywords before you hit Apply.",
                "Generate clean LaTeX resumes on demand."
              ]}
            />
            <p className="mt-6 text-base text-slate-300">
              Resume Tailor combines keyword intelligence with ATS scoring to produce
              a polished, targeted resume that recruiters actually read.
            </p>

            <div className="mt-8 flex flex-wrap items-center gap-4">
              <Link
                href="/login"
                className="rounded-full bg-ember-500 px-6 py-3 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400"
              >
                Get started
              </Link>
              <div className="flex items-center gap-3 text-sm text-slate-400">
                <span className="inline-flex h-2 w-2 rounded-full bg-ember-400" />
                Sign in to upload your CV and paste a job description.
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-6">
            <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-8 shadow-panel backdrop-blur">
              <p className="text-sm uppercase tracking-[0.3em] text-ember-300/80">
                Workflow
              </p>
              <div className="mt-6 space-y-5 text-sm text-slate-300">
                <div className="flex items-start gap-4">
                  <span className="flex h-9 w-9 items-center justify-center rounded-full border border-ember-500/60 bg-ink-950 text-ember-300">
                    1
                  </span>
                  <div>
                    <p className="text-base font-semibold text-white">Upload your CV</p>
                    <p className="mt-1 text-slate-400">
                      Save a base resume and reuse it for every run.
                    </p>
                  </div>
                </div>
                <div className="flex items-start gap-4">
                  <span className="flex h-9 w-9 items-center justify-center rounded-full border border-ember-500/60 bg-ink-950 text-ember-300">
                    2
                  </span>
                  <div>
                    <p className="text-base font-semibold text-white">Paste the job listing</p>
                    <p className="mt-1 text-slate-400">
                      We score it, highlight gaps, and build a change plan.
                    </p>
                  </div>
                </div>
                <div className="flex items-start gap-4">
                  <span className="flex h-9 w-9 items-center justify-center rounded-full border border-ember-500/60 bg-ink-950 text-ember-300">
                    3
                  </span>
                  <div>
                    <p className="text-base font-semibold text-white">Export LaTeX</p>
                    <p className="mt-1 text-slate-400">
                      Get a clean, ATS-friendly resume ready to submit.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-8 shadow-panel backdrop-blur">
              <p className="text-sm uppercase tracking-[0.3em] text-ember-300/80">
                Built for speed
              </p>
              <div className="mt-6 grid gap-4 text-sm text-slate-300">
                {[
                  "ATS score + explanation in under 60 seconds.",
                  "BM25 keyword signals to guide every edit.",
                  "One-click copy into your favorite LaTeX template."
                ].map((item) => (
                  <div
                    key={item}
                    className="rounded-2xl border border-white/10 bg-ink-950/80 px-4 py-3 text-slate-200"
                  >
                    {item}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
