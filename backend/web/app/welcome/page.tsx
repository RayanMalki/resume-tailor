"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

export default function WelcomePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [marking, setMarking] = useState(false);

  useEffect(() => {
    fetch(`${API_BASE_URL}/v1/me`, {
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include",
    })
      .then((r) => {
        if (!r.ok) {
          router.push("/login");
          return null;
        }
        return r.json();
      })
      .then((data) => {
        if (!data) return;
        if (data.onboardingSeen) {
          router.push("/dashboard");
        } else {
          setLoading(false);
        }
      })
      .catch(() => router.push("/login"));
  }, [router]);

  const markSeen = async () => {
    setMarking(true);
    try {
      await fetch(`${API_BASE_URL}/v1/me/onboarding-seen`, {
        method: "POST",
        headers: { "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
      });
    } catch {
      // Best-effort; proceed regardless.
    }
  };

  const handleAddKey = async () => {
    await markSeen();
    router.push("/settings");
  };

  const handleContinue = async () => {
    await markSeen();
    router.push("/dashboard");
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-ink-950 flex items-center justify-center">
        <div className="text-slate-400 text-sm">Loading…</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar />
      <main className="mx-auto max-w-2xl px-4 py-16 sm:px-6">
        <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-8 shadow-panel backdrop-blur sm:p-10">
          <p className="text-xs uppercase tracking-[0.35em] text-ember-300/80">
            Welcome
          </p>
          <h1 className="mt-3 text-2xl font-semibold text-white sm:text-3xl">
            Welcome to Resume Tailor
          </h1>
          <p className="mt-4 text-sm text-slate-400 leading-relaxed">
            Tailor your resume to any job listing in seconds using AI. Here&rsquo;s what
            you need to know before you get started.
          </p>

          {/* Free tier limit */}
          <div className="mt-8 rounded-2xl border border-ember-500/30 bg-ember-950/30 p-5">
            <p className="text-sm font-semibold text-ember-200">
              Free tier: 3 runs per day
            </p>
            <p className="mt-2 text-xs text-slate-400 leading-relaxed">
              Without a personal OpenAI API key, you can run up to{" "}
              <span className="font-semibold text-white">3 analyses per day</span>. The
              limit resets at midnight UTC. This is shared across all devices.
            </p>
          </div>

          {/* API key benefit */}
          <div className="mt-6 rounded-2xl border border-white/10 bg-ink-950/60 p-5">
            <p className="text-sm font-semibold text-white">
              Unlock unlimited runs with your own API key
            </p>
            <p className="mt-2 text-xs text-slate-400 leading-relaxed">
              Add your personal OpenAI API key to remove all daily limits. Your key is
              encrypted with AES-256-GCM and stored securely — it&rsquo;s never returned to
              your browser and is only decrypted in memory during a run.
            </p>

            <div className="mt-4">
              <p className="text-xs font-semibold text-slate-300">
                How to get an OpenAI API key:
              </p>
              <ol className="mt-2 space-y-1 text-xs text-slate-400 list-decimal list-inside leading-relaxed">
                <li>
                  Go to{" "}
                  <a
                    href="https://platform.openai.com/api-keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-ember-300 underline hover:text-ember-200"
                  >
                    platform.openai.com/api-keys
                  </a>
                </li>
                <li>Sign in or create an OpenAI account</li>
                <li>Click &ldquo;Create new secret key&rdquo;</li>
                <li>Copy the key and paste it in Settings</li>
              </ol>
            </div>
          </div>

          {/* CTAs */}
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <button
              onClick={handleAddKey}
              disabled={marking}
              className="flex-1 rounded-full bg-ember-500 px-6 py-3 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400 disabled:opacity-60"
            >
              Add my API key →
            </button>
            <button
              onClick={handleContinue}
              disabled={marking}
              className="flex-1 rounded-full border border-white/10 bg-ink-900 px-6 py-3 text-sm font-semibold text-slate-300 transition hover:border-white/30 hover:text-white disabled:opacity-60"
            >
              Continue without key
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}
