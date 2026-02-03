"use client";

import { useSearchParams, useRouter } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

function JobPageInner() {
  const params = useSearchParams();
  const router = useRouter();
  const resumeId = params.get("resumeId");
  const [jobText, setJobText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!resumeId) {
      router.push("/resume");
    }
  }, [resumeId, router]);

  const handleGenerate = async () => {
    if (!resumeId) return;
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE_URL}/v1/runs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ resumeId, jobText })
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Failed to create run");
      }

      const data = await res.json();
      router.push(`/result/${data.runId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create run");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar showLogout />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-6 py-16">
        <div className="w-full max-w-2xl rounded-[28px] border border-white/10 bg-ink-900/70 p-8 shadow-panel backdrop-blur">
          <h1 className="text-2xl font-semibold text-white">Paste job description</h1>
          <p className="mt-2 text-sm text-slate-400">
            We will tailor your resume to match this job description.
          </p>

          <div className="mt-6">
            <label className="text-sm font-medium text-slate-200">Job description</label>
            <textarea
              value={jobText}
              onChange={(e) => setJobText(e.target.value)}
              rows={12}
              className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2 text-slate-100"
              placeholder="Paste the job description here..."
            />
          </div>

          {error ? <p className="mt-4 text-sm text-rose-300">{error}</p> : null}

          <div className="mt-6 flex justify-end">
            <button
              onClick={handleGenerate}
              disabled={loading || jobText.trim().length === 0}
              className="rounded-full bg-ember-500 px-5 py-2 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400 disabled:opacity-70"
            >
              {loading ? "Generating..." : "Generate"}
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}

export default function JobPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-ink-950" />}>
      <JobPageInner />
    </Suspense>
  );
}
