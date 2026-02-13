"use client";

import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  if (!token) {
    return (
      <div className="rounded-lg border border-rose-500/20 bg-rose-500/10 px-4 py-3">
        <p className="text-sm text-rose-300">Invalid or missing reset token.</p>
        <Link
          href="/forgot-password"
          className="mt-3 inline-block text-sm font-medium text-ember-300 transition hover:text-ember-200"
        >
          Request a new reset link
        </Link>
      </div>
    );
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/auth/reset-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify({ token, password }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => null);
        const msg = data?.error || "Failed to reset password";
        if (msg === "token_expired") throw new Error("This reset link has expired. Please request a new one.");
        if (msg === "token_already_used") throw new Error("This reset link has already been used.");
        if (msg === "invalid_token") throw new Error("Invalid reset link. Please request a new one.");
        throw new Error(msg);
      }

      setSuccess(true);
      setTimeout(() => router.push("/login"), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reset password");
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/10 px-4 py-3">
        <p className="text-sm text-emerald-300">
          Password reset successfully! Redirecting to login&hellip;
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="mt-6 space-y-4">
      <div>
        <label htmlFor="password" className="text-sm font-medium text-slate-200">
          New password
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 placeholder:text-slate-500 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
          placeholder="At least 8 characters"
          required
          minLength={8}
        />
      </div>
      <div>
        <label htmlFor="confirm" className="text-sm font-medium text-slate-200">
          Confirm new password
        </label>
        <input
          id="confirm"
          type="password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 placeholder:text-slate-500 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
          placeholder="Repeat your new password"
          required
          minLength={8}
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
            Resetting&hellip;
          </>
        ) : (
          "Reset password"
        )}
      </button>
    </form>
  );
}

export default function ResetPasswordPage() {
  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-center justify-center px-4 py-10 sm:px-6 sm:py-16">
        <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur sm:p-8">
          <h1 className="text-xl font-semibold text-white sm:text-2xl">Choose a new password</h1>
          <p className="mt-2 text-sm text-slate-400">
            Enter your new password below.
          </p>
          <Suspense fallback={<div className="mt-6 h-40 animate-pulse rounded-lg bg-white/5" />}>
            <ResetPasswordForm />
          </Suspense>
        </div>
      </main>
    </div>
  );
}
