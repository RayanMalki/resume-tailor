import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-ink-950 px-4 text-slate-100">
      <div className="text-center">
        <p className="text-xs uppercase tracking-[0.35em] text-ember-300/80">Error 404</p>
        <h1 className="mt-3 text-4xl font-semibold text-white sm:text-5xl">Page not found</h1>
        <p className="mx-auto mt-4 max-w-md text-sm text-slate-400">
          The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          <Link
            href="/dashboard"
            className="rounded-full bg-ember-500 px-6 py-2.5 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400"
          >
            Go to Dashboard
          </Link>
          <Link
            href="/"
            className="rounded-full border border-white/10 px-6 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
          >
            Home
          </Link>
        </div>
      </div>
    </div>
  );
}
