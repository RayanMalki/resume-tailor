"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";
const REPORT_PROBLEM_MAILTO =
  "mailto:rayanmalki54@gmail.com?subject=Resume%20Tailor%20-%20Problem%20Report";

interface MeResponse {
  userId: string;
  email: string;
  displayName: string;
  avatarUrl: string | null;
  authProvider: string;
  hasApiKey: boolean;
}

export default function TopBar({ showLogout }: { showLogout?: boolean }) {
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [me, setMe] = useState<MeResponse | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!showLogout) return;
    fetch(`${API_BASE_URL}/v1/me`, {
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include",
    })
      .then((r) => r.ok ? r.json() : null)
      .then((data) => { if (data) setMe(data); })
      .catch(() => {});
  }, [showLogout]);

  useEffect(() => {
    if (!dropdownOpen) return;
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [dropdownOpen]);

  const handleLogout = async () => {
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include",
    });
    router.push("/login");
  };

  const initials = me
    ? (me.displayName || me.email || "?")[0].toUpperCase()
    : "?";

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
          <div className="text-base font-semibold text-white">Resume Tailor</div>
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

            {/* Avatar with dropdown */}
            <div className="relative" ref={dropdownRef}>
              <button
                onClick={() => setDropdownOpen((o) => !o)}
                className="flex h-9 w-9 items-center justify-center rounded-full overflow-hidden border border-white/20 bg-ink-900 hover:border-ember-400/60 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500 transition"
                aria-label="User menu"
                aria-expanded={dropdownOpen}
              >
                {me?.avatarUrl ? (
                  <img src={me.avatarUrl} alt={initials} className="h-full w-full object-cover" referrerPolicy="no-referrer" />
                ) : (
                  <span className="text-sm font-bold text-ember-300">{initials}</span>
                )}
              </button>

              {dropdownOpen && (
                <div className="absolute right-0 mt-2 w-40 rounded-xl border border-white/10 bg-ink-900 shadow-xl z-50">
                  <button
                    onClick={() => { setDropdownOpen(false); router.push("/settings"); }}
                    className="w-full rounded-t-xl px-4 py-3 text-left text-sm text-slate-200 hover:bg-white/5 transition"
                  >
                    Settings
                  </button>
                  <a
                    href={REPORT_PROBLEM_MAILTO}
                    onClick={() => setDropdownOpen(false)}
                    className="block w-full px-4 py-3 text-left text-sm text-slate-200 hover:bg-white/5 transition"
                  >
                    Report problem
                  </a>
                  <button
                    onClick={() => { setDropdownOpen(false); handleLogout(); }}
                    className="w-full rounded-b-xl px-4 py-3 text-left text-sm text-ember-300 hover:bg-white/5 transition"
                  >
                    Logout
                  </button>
                </div>
              )}
            </div>
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
              onClick={() => { setMobileOpen(false); router.push("/settings"); }}
              className="w-full rounded-xl border border-white/10 bg-ink-900/60 px-4 py-3 text-left text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
            >
              Settings
            </button>
            <a
              href={REPORT_PROBLEM_MAILTO}
              onClick={() => setMobileOpen(false)}
              className="w-full rounded-xl border border-white/10 bg-ink-900/60 px-4 py-3 text-left text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
            >
              Report problem
            </a>
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
