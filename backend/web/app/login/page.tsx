"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type Mode = "login" | "signup";

export default function LoginPage() {
  const router = useRouter();
  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setLoading(true);
    setError(null);

    const payload: Record<string, string> = { email, password };
    if (mode === "signup") {
      payload.displayName = displayName;
    }

    const endpoint = mode === "login" ? "/v1/auth/login" : "/v1/auth/signup";

    try {
      const res = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Authentication failed");
      }

      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-center justify-center px-4 py-10 sm:px-6 sm:py-16">
        <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur sm:p-8">
          {/* Tab selector */}
          <div className="mb-6 flex gap-2" role="tablist" aria-label="Authentication mode">
            {(["login", "signup"] as Mode[]).map((tab) => (
              <button
                key={tab}
                role="tab"
                aria-selected={mode === tab}
                onClick={() => setMode(tab)}
                className={`flex-1 rounded-full px-4 py-2.5 text-sm font-medium transition focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500 ${
                  mode === tab
                    ? "bg-ember-500 text-ink-950 shadow-glow"
                    : "bg-ink-950 text-slate-300"
                }`}
              >
                {tab === "login" ? "Login" : "Signup"}
              </button>
            ))}
          </div>

          {/* Google OAuth */}
          <button
            type="button"
            onClick={() => {
              window.location.href = `${API_BASE_URL}/v1/auth/google/start?redirect=/dashboard`;
            }}
            className="mb-6 flex w-full items-center justify-center gap-3 rounded-full border border-white/10 bg-ink-950 px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
          >
            <span className="inline-flex h-5 w-5 items-center justify-center rounded-full bg-white text-[11px] font-bold text-ink-950">
              G
            </span>
            Continue with Google
          </button>

          <form onSubmit={handleSubmit} className="space-y-4">
            {mode === "signup" ? (
              <div>
                <label htmlFor="displayName" className="text-sm font-medium text-slate-200">
                  Display name
                </label>
                <input
                  id="displayName"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 placeholder:text-slate-500 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                  required
                />
              </div>
            ) : null}
            <div>
              <label htmlFor="email" className="text-sm font-medium text-slate-200">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 placeholder:text-slate-500 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                required
              />
            </div>
            <div>
              <label htmlFor="password" className="text-sm font-medium text-slate-200">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 placeholder:text-slate-500 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                required
              />
            </div>

            {error ? (
              <div role="alert" className="rounded-lg border border-rose-500/20 bg-rose-500/10 px-3 py-2">
                <p className="text-sm text-rose-300">{error}</p>
              </div>
            ) : null}

            <button
              type="submit"
              disabled={loading}
              className="flex w-full items-center justify-center gap-2 rounded-full bg-ember-500 px-4 py-2.5 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400 disabled:opacity-70 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500 focus-visible:ring-offset-2 focus-visible:ring-offset-ink-900"
            >
              {loading ? (
                <>
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-ink-950/30 border-t-ink-950" />
                  Please wait&hellip;
                </>
              ) : mode === "login" ? "Login" : "Create account"}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}
