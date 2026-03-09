"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";
const REPORT_PROBLEM_MAILTO =
  "mailto:rayanmalki54@gmail.com?subject=Resume%20Tailor%20-%20Problem%20Report";

type UserData = {
  email: string;
  displayName: string;
  avatarUrl: string | null;
};

type EditorialNavProps = {
  mode: "public" | "private";
};

export default function EditorialNav({ mode }: EditorialNavProps) {
  const router = useRouter();

  const [mobileOpen, setMobileOpen] = useState(false);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [me, setMe] = useState<UserData | null>(null);

  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (mode !== "private") return;
    fetch(`${API_BASE_URL}/v1/me`, {
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include",
    })
      .then((response) => (response.ok ? response.json() : null))
      .then((data: UserData | null) => {
        if (data) setMe(data);
      })
      .catch(() => {});
  }, [mode]);

  useEffect(() => {
    if (!dropdownOpen) return;
    const onClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", onClickOutside);
    return () => document.removeEventListener("mousedown", onClickOutside);
  }, [dropdownOpen]);

  const logout = async () => {
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
    <header className="rt-topbar px-4 py-3 sm:px-5 sm:py-4">
      <nav
        className="flex items-center justify-between gap-4"
        aria-label="Primary navigation"
      >
        <div className="flex min-w-0 items-center gap-3">
          <span
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-[rgba(49,45,35,0.28)] bg-[#f2b66f] text-xs font-bold text-[#151513]"
            aria-hidden="true"
          >
            RT
          </span>
          <div className="min-w-0">
            <p className="truncate font-grotesk text-xl font-semibold leading-none text-[var(--rt-ink-900)]">
              Resume Tailor
            </p>
            <p className="rt-label mt-1 hidden sm:block">built by a person, for people</p>
          </div>
        </div>

        {mode === "public" ? (
          <Link
            href="/login"
            className="rt-btn-primary hidden px-5 py-2 text-xs font-semibold uppercase tracking-[0.2em] md:inline-flex"
          >
            Login
          </Link>
        ) : (
          <div className="hidden items-center gap-2 md:flex">
            <button
              onClick={() => router.push("/dashboard")}
              className="rt-btn-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.2em]"
              aria-label="Go to dashboard"
            >
              Dashboard
            </button>

            <div className="relative" ref={dropdownRef}>
              <button
                onClick={() => setDropdownOpen((open) => !open)}
                className="inline-flex h-9 w-9 items-center justify-center overflow-hidden rounded-full border border-[rgba(67,63,52,0.25)] bg-white/70 text-sm font-bold text-[var(--rt-ink-700)]"
                aria-label="Open user menu"
                aria-expanded={dropdownOpen}
              >
                {me?.avatarUrl ? (
                  <img
                    src={me.avatarUrl}
                    alt={initials}
                    className="h-full w-full object-cover"
                    referrerPolicy="no-referrer"
                  />
                ) : (
                  <span>{initials}</span>
                )}
              </button>

              {dropdownOpen ? (
                <div className="absolute right-0 top-full z-50 mt-2 w-44 overflow-hidden rounded-2xl border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.94)] text-[var(--rt-ink-700)] shadow-[var(--rt-shadow-soft)]">
                  <button
                    onClick={() => {
                      setDropdownOpen(false);
                      router.push("/settings");
                    }}
                    className="w-full px-4 py-3 text-left text-sm transition hover:bg-[#f2eee6]"
                  >
                    Settings
                  </button>
                  <a
                    href={REPORT_PROBLEM_MAILTO}
                    onClick={() => setDropdownOpen(false)}
                    className="block px-4 py-3 text-sm transition hover:bg-[#f2eee6]"
                  >
                    Report problem
                  </a>
                  <button
                    onClick={() => {
                      setDropdownOpen(false);
                      logout();
                    }}
                    className="w-full px-4 py-3 text-left text-sm text-[var(--rt-accent-strong)] transition hover:bg-[#f9ece8]"
                  >
                    Logout
                  </button>
                </div>
              ) : null}
            </div>
          </div>
        )}

        <button
          onClick={() => setMobileOpen((open) => !open)}
          className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/60 text-[var(--rt-ink-700)] md:hidden"
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
          aria-expanded={mobileOpen}
        >
          {mobileOpen ? (
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          ) : (
            <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          )}
        </button>
      </nav>

      {mobileOpen ? (
        <div className="mt-3 border-t border-[var(--rt-stroke)] pt-3 md:hidden">
          <div className="grid gap-2">
            {mode === "public" ? (
              <Link
                href="/login"
                onClick={() => setMobileOpen(false)}
                className="rt-btn-primary px-4 py-3 text-center text-xs font-semibold uppercase tracking-[0.2em]"
              >
                Login
              </Link>
            ) : (
              <>
                <button
                  onClick={() => {
                    setMobileOpen(false);
                    router.push("/dashboard");
                  }}
                  className="rt-btn-primary px-4 py-3 text-xs font-semibold uppercase tracking-[0.2em]"
                >
                  Dashboard
                </button>
                <button
                  onClick={() => {
                    setMobileOpen(false);
                    router.push("/settings");
                  }}
                  className="rt-btn-secondary px-4 py-3 text-left text-xs font-semibold uppercase tracking-[0.2em]"
                >
                  Settings
                </button>
                <a
                  href={REPORT_PROBLEM_MAILTO}
                  onClick={() => setMobileOpen(false)}
                  className="rt-btn-secondary px-4 py-3 text-left text-xs font-semibold uppercase tracking-[0.2em]"
                >
                  Report problem
                </a>
                <button
                  onClick={() => {
                    setMobileOpen(false);
                    logout();
                  }}
                  className="rounded-full border border-[rgba(192,79,49,0.25)] bg-[rgba(255,240,235,0.85)] px-4 py-3 text-left text-xs font-semibold uppercase tracking-[0.2em] text-[var(--rt-accent-strong)]"
                >
                  Logout
                </button>
              </>
            )}
          </div>
        </div>
      ) : null}
    </header>
  );
}
