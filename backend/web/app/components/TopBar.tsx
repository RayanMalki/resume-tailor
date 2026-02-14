"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

export default function TopBar({ showLogout }: { showLogout?: boolean }) {
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleLogout = async () => {
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include"
    });
    router.push("/login");
  };

  return (
    <header className="w-full border-b border-white/10 bg-ink-950/70 backdrop-blur">
      <nav className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 py-3 sm:px-6 sm:py-4" aria-label="Main navigation">
        {/* Logo */}
        <a
          href="/dashboard"
          onClick={(e) => { e.preventDefault(); router.push("/dashboard"); }}
          className="flex items-center gap-3 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500 focus-visible:ring-offset-2 focus-visible:ring-offset-ink-950 rounded-lg"
          aria-label="Resume Tailor — go to dashboard"
        >
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-ember-500/60 bg-ink-900 text-sm font-semibold text-ember-300 shadow-glow">
            RT
          </span>
          <div>
            <div className="text-base font-semibold text-white">Resume Tailor</div>
            <div className="text-[11px] uppercase tracking-[0.3em] text-ember-300/80">
              dashboard
            </div>
          </div>
        </a>

        {/* Desktop nav */}
        {showLogout ? (
          <div className="hidden items-center gap-2 sm:flex">
            <button
              onClick={() => router.push("/dashboard")}
              className="rounded-full border border-white/10 bg-ink-900/60 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
            >
              Dashboard
            </button>
            <button
              onClick={handleLogout}
              className="rounded-full border border-ember-500/60 bg-ink-900/60 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-ember-200 transition hover:border-ember-400 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
            >
              Logout
            </button>
          </div>
        ) : null}

        {/* Mobile hamburger */}
        {showLogout ? (
          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 text-slate-200 transition hover:border-ember-400/60 sm:hidden focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
            aria-expanded={mobileOpen}
          >
            {mobileOpen ? (
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            ) : (
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            )}
          </button>
        ) : null}
      </nav>

      {/* Mobile menu dropdown */}
      {showLogout && mobileOpen ? (
        <div className="border-t border-white/10 px-4 py-3 sm:hidden">
          <div className="flex flex-col gap-2">
            <button
              onClick={() => { setMobileOpen(false); router.push("/dashboard"); }}
              className="w-full rounded-xl border border-white/10 bg-ink-900/60 px-4 py-3 text-left text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
            >
              Dashboard
            </button>
            <button
              onClick={() => { setMobileOpen(false); handleLogout(); }}
              className="w-full rounded-xl border border-ember-500/60 bg-ink-900/60 px-4 py-3 text-left text-sm font-semibold text-ember-200 transition hover:border-ember-400 hover:text-white"
            >
              Logout
            </button>
          </div>
        </div>
      ) : null}
    </header>
  );
}
