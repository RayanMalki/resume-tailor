"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import EditorialNav from "../components/EditorialNav";

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
      <div className="rt-canvas">
        <div className="rt-shell flex items-center justify-center min-h-screen">
          <div className="text-sm text-[var(--rt-ink-500)]">Loading…</div>
        </div>
      </div>
    );
  }

  return (
    <div className="rt-canvas">
      <EditorialNav mode="private" />
      <div className="rt-shell">
        <main className="mt-4 space-y-4 sm:mt-6 sm:space-y-6">
          {/* Hero panel */}
          <section
            className="rt-panel rt-fade-up p-5 sm:p-7"
            style={{ animationDelay: "40ms" }}
          >
            <p className="rt-label">Welcome</p>
            <h1 className="mt-2 font-grotesk text-4xl font-semibold text-[var(--rt-ink-900)] sm:text-5xl">
              Welcome to Resume Tailor
            </h1>
            <p className="font-serif-display mt-2 text-xl text-[var(--rt-ink-700)]">
              Tailor your resume to any job listing in seconds using AI. Here&rsquo;s
              what you need to know before you get started.
            </p>
          </section>

          {/* Free tier info */}
          <section
            className="rt-panel-muted rt-fade-up p-5 sm:p-7"
            style={{ animationDelay: "120ms" }}
          >
            <p className="rt-label">Free tier</p>
            <h2 className="mt-2 font-grotesk text-lg font-semibold text-[var(--rt-ink-900)]">
              3 runs per day
            </h2>
            <p className="mt-2 text-sm text-[var(--rt-ink-700)] leading-relaxed">
              Without a personal OpenAI API key, you can run up to{" "}
              <span className="font-semibold text-[var(--rt-ink-900)]">3 analyses per day</span>.
              The limit resets at midnight UTC. This is shared across all devices.
            </p>
          </section>

          {/* API key info */}
          <section
            className="rt-panel rt-fade-up p-5 sm:p-7"
            style={{ animationDelay: "160ms" }}
          >
            <p className="rt-label">API key</p>
            <h2 className="mt-2 font-grotesk text-lg font-semibold text-[var(--rt-ink-900)]">
              Unlock unlimited runs with your own API key
            </h2>
            <p className="mt-2 text-sm text-[var(--rt-ink-700)] leading-relaxed">
              Add your personal OpenAI API key to remove all daily limits. Your key is
              encrypted with AES-256-GCM and stored securely — it&rsquo;s never returned to
              your browser and is only decrypted in memory during a run.
            </p>

            <div className="mt-4">
              <p className="text-sm font-semibold text-[var(--rt-ink-900)]">
                How to get an OpenAI API key:
              </p>
              <ol className="mt-2 space-y-1 text-sm text-[var(--rt-ink-700)] list-decimal list-inside leading-relaxed">
                <li>
                  Go to{" "}
                  <a
                    href="https://platform.openai.com/api-keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-[var(--rt-accent)] underline hover:opacity-80"
                  >
                    platform.openai.com/api-keys
                  </a>
                </li>
                <li>Sign in or create an OpenAI account</li>
                <li>Click &ldquo;Create new secret key&rdquo;</li>
                <li>Copy the key and paste it in Settings</li>
              </ol>
            </div>
          </section>

          {/* CTAs */}
          <section
            className="rt-panel rt-fade-up p-5 sm:p-7"
            style={{ animationDelay: "200ms" }}
          >
            <div className="flex flex-col gap-3 sm:flex-row">
              <button
                onClick={handleAddKey}
                disabled={marking}
                className="rt-btn-primary flex-1 disabled:opacity-60"
              >
                Add my API key →
              </button>
              <button
                onClick={handleContinue}
                disabled={marking}
                className="rt-btn-secondary flex-1 disabled:opacity-60"
              >
                Continue without key
              </button>
            </div>
          </section>
        </main>
      </div>
    </div>
  );
}
