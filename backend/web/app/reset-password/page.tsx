"use client";

import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import EditorialNav from "../components/EditorialNav";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

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
      <div className="mt-6 rounded-lg border border-[rgba(190,76,47,0.2)] bg-[rgba(255,236,229,0.85)] px-4 py-3">
        <p className="text-sm text-[var(--rt-accent-strong)]">Invalid or missing reset token.</p>
        <Link
          href="/forgot-password"
          className="mt-3 inline-block text-sm font-medium text-[var(--rt-accent)] transition hover:text-[var(--rt-accent-strong)]"
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
      <div className="mt-6 rounded-lg border border-[rgba(41,137,109,0.25)] bg-[rgba(229,246,239,0.78)] px-4 py-3">
        <p className="text-sm text-[rgba(28,112,87,0.92)]">
          Password reset successfully! Redirecting to login&hellip;
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="mt-6 space-y-4">
      <div>
        <label htmlFor="password" className="text-sm font-medium text-[var(--rt-ink-700)]">
          New password
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-[var(--rt-accent)] focus:outline-none focus:ring-1 focus:ring-[rgba(190,76,47,0.3)]"
          placeholder="At least 8 characters"
          required
          minLength={8}
        />
      </div>
      <div>
        <label htmlFor="confirm" className="text-sm font-medium text-[var(--rt-ink-700)]">
          Confirm new password
        </label>
        <input
          id="confirm"
          type="password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-[var(--rt-accent)] focus:outline-none focus:ring-1 focus:ring-[rgba(190,76,47,0.3)]"
          placeholder="Repeat your new password"
          required
          minLength={8}
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
            Resetting&hellip;
          </>
        ) : "Reset password"}
      </button>
    </form>
  );
}

export default function ResetPasswordPage() {
  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="public" />
        <main className="flex min-h-[calc(100vh-5rem)] items-center justify-center py-10">
          <div className="rt-panel w-full max-w-md p-6 sm:p-8">
            <p className="rt-label">Account</p>
            <h1 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
              Choose a new password
            </h1>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              Enter your new password below.
            </p>
            <Suspense fallback={<div className="mt-6 h-40 animate-pulse rounded-lg bg-[rgba(81,73,62,0.08)]" />}>
              <ResetPasswordForm />
            </Suspense>
          </div>
        </main>
      </div>
    </div>
  );
}
