"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import EditorialNav from "../components/EditorialNav";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

type VerifyState = "loading" | "success" | "error";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [state, setState] = useState<VerifyState>(token ? "loading" : "error");
  const [errorMsg, setErrorMsg] = useState(token ? "" : "Missing verification token.");

  const [resending, setResending] = useState(false);
  const [resendResult, setResendResult] = useState<string | null>(null);

  useEffect(() => {
    if (!token) return;

    const verify = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/auth/verify?token=${encodeURIComponent(token)}`, {
          credentials: "include",
        });

        if (res.ok) {
          setState("success");
          return;
        }

        const data = await res.json().catch(() => null);
        const code = data?.error || "";

        if (code === "token_expired") {
          setErrorMsg("This verification link has expired. Please request a new one.");
        } else if (code === "token_already_used") {
          setErrorMsg("This email has already been verified. You can log in.");
        } else if (code === "invalid_token") {
          setErrorMsg("Invalid verification link. Please request a new one.");
        } else {
          setErrorMsg("Verification failed. Please try again.");
        }
        setState("error");
      } catch {
        setErrorMsg("Network error. Please try again.");
        setState("error");
      }
    };

    verify();
  }, [token]);

  const handleResend = async () => {
    setResending(true);
    setResendResult(null);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/auth/resend-verification`, {
        method: "POST",
        headers: { "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
      });

      if (res.status === 401) {
        setResendResult("You need to be logged in to resend the verification email.");
        return;
      }

      if (res.ok) {
        setResendResult("Verification email sent! Check your inbox.");
      } else {
        setResendResult("Failed to resend. Please try again later.");
      }
    } catch {
      setResendResult("Network error. Please try again.");
    } finally {
      setResending(false);
    }
  };

  return (
    <>
      {state === "loading" && (
        <div className="mt-6 flex flex-col items-center gap-4">
          <div className="h-10 w-10 animate-spin rounded-full border-2 border-[rgba(190,76,47,0.3)] border-t-[var(--rt-accent)]" />
          <p className="text-sm text-[var(--rt-ink-700)]">Verifying your email&hellip;</p>
        </div>
      )}

      {state === "success" && (
        <div className="mt-6 rounded-lg border border-[rgba(41,137,109,0.25)] bg-[rgba(229,246,239,0.78)] px-4 py-3">
          <p className="text-sm text-[rgba(28,112,87,0.92)]">
            Your email has been verified successfully!
          </p>
          <Link
            href="/login"
            className="rt-btn-primary mt-3 inline-flex px-5 py-2 text-sm font-semibold"
          >
            Go to login
          </Link>
        </div>
      )}

      {state === "error" && (
        <div className="mt-6 space-y-4">
          <div className="rounded-lg border border-[rgba(190,76,47,0.2)] bg-[rgba(255,236,229,0.85)] px-4 py-3">
            <p className="text-sm text-[var(--rt-accent-strong)]">{errorMsg}</p>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <button
              onClick={handleResend}
              disabled={resending}
              className="flex items-center justify-center gap-2 rounded-full border border-[rgba(67,63,52,0.25)] bg-white/70 px-4 py-2.5 text-sm font-semibold text-[var(--rt-ink-700)] transition hover:border-[var(--rt-accent)] hover:text-[var(--rt-accent)] disabled:opacity-70"
            >
              {resending ? (
                <>
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-[rgba(67,63,52,0.2)] border-t-[var(--rt-ink-700)]" />
                  Sending&hellip;
                </>
              ) : "Resend verification email"}
            </button>
            <Link
              href="/login"
              className="flex items-center justify-center rounded-full border border-[rgba(67,63,52,0.25)] bg-white/70 px-4 py-2.5 text-sm font-semibold text-[var(--rt-ink-700)] transition hover:border-[var(--rt-accent)] hover:text-[var(--rt-accent)]"
            >
              Back to login
            </Link>
          </div>

          {resendResult ? (
            <p className="text-sm text-[var(--rt-ink-700)]">{resendResult}</p>
          ) : null}
        </div>
      )}
    </>
  );
}

export default function VerifyPage() {
  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="public" />
        <main className="flex min-h-[calc(100vh-5rem)] items-center justify-center py-10">
          <div className="rt-panel w-full max-w-md p-6 sm:p-8">
            <p className="rt-label">Account</p>
            <h1 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
              Email verification
            </h1>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              Confirming your email address&hellip;
            </p>
            <Suspense fallback={<div className="mt-6 h-20 animate-pulse rounded-lg bg-[rgba(81,73,62,0.08)]" />}>
              <VerifyEmailContent />
            </Suspense>
          </div>
        </main>
      </div>
    </div>
  );
}
