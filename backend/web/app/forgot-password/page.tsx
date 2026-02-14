"use client";

import { useState } from "react";
import Link from "next/link";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE_URL}/v1/auth/forgot-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify({ email }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => null);
        throw new Error(data?.error || "Something went wrong");
      }

      setSent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-center justify-center px-4 py-10 sm:px-6 sm:py-16">
        <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur sm:p-8">
          <h1 className="text-xl font-semibold text-white sm:text-2xl">Reset your password</h1>
          <p className="mt-2 text-sm text-slate-400">
            Enter the email address associated with your account and we&apos;ll send you a link to reset your password.
          </p>

          {sent ? (
            <div className="mt-6 rounded-lg border border-emerald-500/20 bg-emerald-500/10 px-4 py-3">
              <p className="text-sm text-emerald-300">
                If an account with that email exists, we&apos;ve sent a password reset link. Check your inbox.
              </p>
              <Link
                href="/login"
                className="mt-3 inline-block text-sm font-medium text-ember-300 transition hover:text-ember-200"
              >
                &larr; Back to login
              </Link>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="mt-6 space-y-4">
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
                  placeholder="you@example.com"
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
                    Sending&hellip;
                  </>
                ) : (
                  "Send reset link"
                )}
              </button>

              <div className="text-center">
                <Link
                  href="/login"
                  className="text-sm text-slate-400 transition hover:text-ember-300"
                >
                  &larr; Back to login
                </Link>
              </div>
            </form>
          )}
        </div>
      </main>
    </div>
  );
}
