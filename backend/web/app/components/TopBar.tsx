"use client";

import { useRouter } from "next/navigation";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export default function TopBar({ showLogout }: { showLogout?: boolean }) {
  const router = useRouter();

  const handleLogout = async () => {
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      credentials: "include"
    });
    router.push("/login");
  };

  return (
    <header className="w-full border-b border-white/10 bg-ink-950/70 backdrop-blur">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-4">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-ember-500/60 bg-ink-900 text-sm font-semibold text-ember-300 shadow-glow">
            RT
          </span>
          <div>
            <div className="text-base font-semibold text-white">Resume Tailor</div>
            <div className="text-[11px] uppercase tracking-[0.3em] text-ember-300/80">
              dashboard
            </div>
          </div>
        </div>
        {showLogout ? (
          <div className="flex items-center gap-2">
            <button
              onClick={() => router.push("/dashboard")}
              className="rounded-full border border-white/10 bg-ink-900/60 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.2em] text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
            >
              Dashboard
            </button>
            <button
              onClick={handleLogout}
              className="rounded-full border border-ember-500/60 bg-ink-900/60 px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.2em] text-ember-200 transition hover:border-ember-400 hover:text-white"
            >
              Logout
            </button>
          </div>
        ) : null}
      </div>
    </header>
  );
}
