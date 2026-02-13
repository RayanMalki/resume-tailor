"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type VerifyState = "loading" | "success" | "error";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [state, setState] = useState<VerifyState>(token ? "loading" : "error");
  const [errorMsg, setErrorMsg] = useState(token ? "" : "Missing verification token.");

  // Resend state
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
          <div className="h-10 w-10 animate-spin rounded-full border-2 border-ember-500/30 border-t-ember-500" />
          <p className="text-sm text-slate-400">Verifying your email&hellip;</p>
        </div>
      )}

      {state === "success" && (
        <div className="mt-6 rounded-lg border border-emerald-500/20 bg-emerald-500/10 px-4 py-3">
          <p className="text-sm text-emerald-300">
            Your email has been verified successfully!
          </p>
          <Link
            href="/login"
            className="mt-3 inline-block rounded-full bg-ember-500 px-5 py-2 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400"
          >
            Go to login
          </Link>
        </div>
      )}

      {state === "error" && (
        <div className="mt-6 space-y-4">
          <div className="rounded-lg border border-rose-500/20 bg-rose-500/10 px-4 py-3">
            <p className="text-sm text-rose-300">{errorMsg}</p>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <button
              onClick={handleResend}
              disabled={resending}
              className="flex items-center justify-center gap-2 rounded-full border border-white/10 px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200 disabled:opacity-70"
            >
              {resending ? (
                <>
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-slate-200/30 border-t-slate-200" />
                  Sending&hellip;
                </>
              ) : (
                "Resend verification email"
              )}
            </button>
            <Link
              href="/login"
              className="flex items-center justify-center rounded-full border border-white/10 px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200"
            >
              Back to login
            </Link>
          </div>

          {resendResult ? (
            <p className="text-sm text-slate-400">{resendResult}</p>
          ) : null}
        </div>
      )}
    </>
  );
}

export default function VerifyPage() {
  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-center justify-center px-4 py-10 sm:px-6 sm:py-16">
        <div className="w-full max-w-md rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur sm:p-8">
          <h1 className="text-xl font-semibold text-white sm:text-2xl">Email verification</h1>
          <p className="mt-2 text-sm text-slate-400">
            Confirming your email address&hellip;
          </p>
          <Suspense fallback={<div className="mt-6 h-20 animate-pulse rounded-lg bg-white/5" />}>
            <VerifyEmailContent />
          </Suspense>
        </div>
      </main>
    </div>
  );
}
