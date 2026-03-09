"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import EditorialNav from "../components/EditorialNav";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

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
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="public" />
        <main className="flex min-h-[calc(100vh-5rem)] items-center justify-center py-10">
          <div className="rt-panel w-full max-w-md p-6 sm:p-8">
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
                      ? "bg-[var(--rt-accent)] text-[#fff8f3]"
                      : "bg-white/70 border border-[rgba(67,63,52,0.25)] text-[var(--rt-ink-700)]"
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
              className="mb-6 flex w-full items-center justify-center gap-3 rounded-full border border-[rgba(67,63,52,0.25)] bg-white/70 px-4 py-2.5 text-sm font-semibold text-[var(--rt-ink-900)] transition hover:border-ember-400/60 hover:text-[var(--rt-accent)] focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
            >
              <span className="inline-flex h-5 w-5 items-center justify-center rounded-full bg-white text-[11px] font-bold text-ink-950">
                G
              </span>
              Continue with Google
            </button>

            <form onSubmit={handleSubmit} className="space-y-4">
              {mode === "signup" ? (
                <div>
                  <label htmlFor="displayName" className="text-sm font-medium text-[var(--rt-ink-700)]">
                    Display name
                  </label>
                  <input
                    id="displayName"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                    required
                  />
                </div>
              ) : null}
              <div>
                <label htmlFor="email" className="text-sm font-medium text-[var(--rt-ink-700)]">
                  Email
                </label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                  required
                />
              </div>
              <div>
                <label htmlFor="password" className="text-sm font-medium text-[var(--rt-ink-700)]">
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                  required
                />
              </div>

              {mode === "login" ? (
                <div className="text-right">
                  <Link
                    href="/forgot-password"
                    className="text-xs text-[var(--rt-ink-500)] transition hover:text-[var(--rt-accent)]"
                  >
                    Forgot password?
                  </Link>
                </div>
              ) : null}

              {error ? (
                <div role="alert" className="rounded-lg border border-[rgba(190,76,47,0.2)] bg-[rgba(255,236,229,0.85)] px-3 py-2">
                  <p className="text-sm text-[var(--rt-accent-strong)]">{error}</p>
                </div>
              ) : null}

              <button
                type="submit"
                disabled={loading}
                className="rt-btn-primary flex w-full items-center justify-center gap-2 px-4 py-2.5 text-sm font-semibold disabled:opacity-70 focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--rt-accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-white"
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
    </div>
  );
}
