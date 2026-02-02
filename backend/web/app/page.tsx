import Link from "next/link";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-slate-50">
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
        <div className="text-lg font-semibold text-slate-900">Resume Tailor</div>
        <Link
          href="/login"
          className="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-700 hover:border-slate-300"
        >
          Login
        </Link>
      </header>

      <main className="mx-auto w-full max-w-6xl px-6 pb-16">
        <section className="rounded-3xl bg-white p-10 shadow-card">
          <div className="max-w-2xl">
            <p className="text-sm font-medium uppercase tracking-wide text-slate-500">
              Welcome
            </p>
            <h1 className="mt-3 text-3xl font-semibold text-slate-900 sm:text-4xl">
              Tailor your resume to every job in minutes.
            </h1>
            <p className="mt-4 text-base text-slate-600">
              Resume Tailor takes your existing CV and the job description you want,
              then produces a focused, ATS-friendly LaTeX resume that highlights the
              most relevant experience and skills.
            </p>

            <div className="mt-6 flex flex-wrap gap-3">
              <Link
                href="/login"
                className="rounded-lg bg-slate-900 px-5 py-2.5 text-sm font-medium text-white hover:bg-slate-800"
              >
                Get started
              </Link>
              <span className="text-sm text-slate-500 self-center">
                Sign in to upload your CV and paste a job description.
              </span>
            </div>
          </div>
        </section>

        <section className="mt-10 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl bg-white p-6 shadow-card">
            <p className="text-sm font-semibold text-slate-800">1. Upload your CV</p>
            <p className="mt-2 text-sm text-slate-600">
              Add your current resume once, then reuse it for every job.
            </p>
          </div>
          <div className="rounded-2xl bg-white p-6 shadow-card">
            <p className="text-sm font-semibold text-slate-800">2. Paste the job listing</p>
            <p className="mt-2 text-sm text-slate-600">
              Drop the job description so we can match your experience to it.
            </p>
          </div>
          <div className="rounded-2xl bg-white p-6 shadow-card">
            <p className="text-sm font-semibold text-slate-800">3. Get tailored LaTeX</p>
            <p className="mt-2 text-sm text-slate-600">
              Receive a clean, one-page LaTeX resume ready for export and edits.
            </p>
          </div>
        </section>
      </main>
    </div>
  );
}
