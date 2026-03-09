"use client";

import { useState } from "react";
import Link from "next/link";
import EditorialNav from "../components/EditorialNav";

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
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="public" />
        <main className="flex min-h-[calc(100vh-5rem)] items-center justify-center py-10">
          <div className="rt-panel w-full max-w-md p-6 sm:p-8">
            <p className="rt-label">Account</p>
            <h1 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
              Reset your password
            </h1>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              Enter the email address associated with your account and we&apos;ll send you a reset link.
            </p>

            {sent ? (
              <div className="mt-6 rounded-lg border border-[rgba(41,137,109,0.25)] bg-[rgba(229,246,239,0.78)] px-4 py-3">
                <p className="text-sm text-[rgba(28,112,87,0.92)]">
                  If an account with that email exists, we&apos;ve sent a password reset link. Check your inbox.
                </p>
                <Link
                  href="/login"
                  className="mt-3 inline-block text-sm font-medium text-[var(--rt-accent)] transition hover:text-[var(--rt-accent-strong)]"
                >
                  &larr; Back to login
                </Link>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="mt-6 space-y-4">
                <div>
                  <label htmlFor="email" className="text-sm font-medium text-[var(--rt-ink-700)]">
                    Email
                  </label>
                  <input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-[var(--rt-accent)] focus:outline-none focus:ring-1 focus:ring-[rgba(190,76,47,0.3)]"
                    placeholder="you@example.com"
                    required
                  />
                </div>

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
                      <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                      Sending&hellip;
                    </>
                  ) : "Send reset link"}
                </button>

                <div className="text-center">
                  <Link
                    href="/login"
                    className="text-sm text-[var(--rt-ink-500)] transition hover:text-[var(--rt-accent)]"
                  >
                    &larr; Back to login
                  </Link>
                </div>
              </form>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}
