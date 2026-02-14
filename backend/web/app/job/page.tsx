"use client";

import { useSearchParams, useRouter } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

type Discipline =
  | "mechanical"
  | "electrical"
  | "industrial_logistics"
  | "aerospace"
  | "it_software";

const disciplineOptions: Array<{ value: Discipline; label: string }> = [
  { value: "mechanical", label: "Mechanical Engineering" },
  { value: "electrical", label: "Electrical Engineering" },
  { value: "industrial_logistics", label: "Industrial / Operations / Logistics" },
  { value: "aerospace", label: "Aerospace Engineering" },
  { value: "it_software", label: "IT / Software Engineering" },
];

function labelForDiscipline(value: string | null | undefined): string {
  if (!value) return "Unknown";
  const found = disciplineOptions.find((option) => option.value === value);
  return found ? found.label : value;
}

function JobPageInner() {
  const params = useSearchParams();
  const router = useRouter();
  const resumeId = params.get("resumeId");
  const [jobText, setJobText] = useState("");
  const [detectedDiscipline, setDetectedDiscipline] = useState<Discipline | null>(null);
  const [detectedConfidence, setDetectedConfidence] = useState<number>(0);
  const [lowConfidence, setLowConfidence] = useState(false);
  const [detecting, setDetecting] = useState(false);
  const [disciplineOverride, setDisciplineOverride] = useState<"" | Discipline>("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!resumeId) {
      router.push("/resume");
    }
  }, [resumeId, router]);

  useEffect(() => {
    if (!resumeId) return;
    if (jobText.trim().length < 80) {
      setDetectedDiscipline(null);
      setDetectedConfidence(0);
      setLowConfidence(false);
      return;
    }

    const controller = new AbortController();
    const timeout = setTimeout(async () => {
      setDetecting(true);
      try {
        const res = await fetch(`${API_BASE_URL}/v1/disciplines/detect`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
          credentials: "include",
          signal: controller.signal,
          body: JSON.stringify({
            resumeId,
            jobText,
          }),
        });
        if (!res.ok) return;
        const data = await res.json();
        if (typeof data.discipline === "string") {
          setDetectedDiscipline(data.discipline as Discipline);
        }
        setDetectedConfidence(typeof data.confidence === "number" ? data.confidence : 0);
        setLowConfidence(Boolean(data.lowConfidence));
      } catch {
        // Non-blocking: discipline detect should not block run creation.
      } finally {
        setDetecting(false);
      }
    }, 550);

    return () => {
      clearTimeout(timeout);
      controller.abort();
    };
  }, [resumeId, jobText]);

  const handleGenerate = async () => {
    if (!resumeId) return;
    setLoading(true);
    setError(null);

    try {
      const payload: Record<string, unknown> = {
        resumeId,
        jobText,
      };
      if (disciplineOverride) {
        payload.disciplineOverride = disciplineOverride;
      }

      const res = await fetch(`${API_BASE_URL}/v1/runs`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify(payload),
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
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-4 py-8 sm:px-6 sm:py-16">
        <div className="w-full max-w-3xl rounded-[28px] border border-white/10 bg-ink-900/70 p-5 shadow-panel backdrop-blur sm:p-8">
          <h1 className="text-xl font-semibold text-white sm:text-2xl">Paste job description</h1>
          <p className="mt-2 text-sm text-slate-400">
            We&apos;ll tailor your resume to match this job description.
          </p>

          <div className="mt-6">
            <label htmlFor="job-text" className="text-sm font-medium text-slate-200">Job description</label>
            <textarea
              id="job-text"
              value={jobText}
              onChange={(e) => setJobText(e.target.value)}
              rows={10}
              className="mt-1 w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-slate-100 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
              placeholder="Paste the job description here..."
            />
          </div>

          <div className="mt-4 rounded-xl border border-white/10 bg-ink-950/40 p-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-xs uppercase tracking-[0.18em] text-slate-400">Detected discipline</p>
              {detecting ? <p className="text-xs text-slate-500">Detecting&hellip;</p> : null}
            </div>
            <p className="mt-1 text-sm text-slate-200">
              {detectedDiscipline ? labelForDiscipline(detectedDiscipline) : "Add more job details to detect"}
            </p>
            {detectedDiscipline ? (
              <p className={`mt-1 text-xs ${lowConfidence ? "text-amber-300" : "text-slate-400"}`}>
                Confidence: {Math.round(Math.max(0, Math.min(1, detectedConfidence)) * 100)}%
                {lowConfidence ? " (low confidence)" : ""}
              </p>
            ) : null}

            <label htmlFor="discipline-override" className="mt-4 block text-xs font-medium uppercase tracking-[0.15em] text-slate-400">
              Override discipline (optional)
            </label>
            <select
              id="discipline-override"
              value={disciplineOverride}
              onChange={(e) => setDisciplineOverride(e.target.value as "" | Discipline)}
              className="mt-1 w-full rounded-lg border border-white/10 bg-ink-950 px-3 py-2 text-sm text-slate-100 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
            >
              <option value="">Use detected discipline</option>
              {disciplineOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          {error ? (
            <div role="alert" className="mt-4 rounded-lg border border-rose-500/20 bg-rose-500/10 px-3 py-2">
              <p className="text-sm text-rose-300">{error}</p>
            </div>
          ) : null}

          <div className="mt-6 flex justify-end">
            <button
              onClick={handleGenerate}
              disabled={loading || jobText.trim().length === 0}
              className="flex items-center gap-2 rounded-full bg-ember-500 px-5 py-2.5 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400 disabled:opacity-70 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500 focus-visible:ring-offset-2 focus-visible:ring-offset-ink-900"
            >
              {loading ? (
                <>
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-ink-950/30 border-t-ink-950" />
                  Generating&hellip;
                </>
              ) : "Generate"}
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
