"use client";

import { useSearchParams, useRouter } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import EditorialNav from "../components/EditorialNav";

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
          body: JSON.stringify({ resumeId, jobText }),
        });
        if (!res.ok) return;
        const data = await res.json();
        if (typeof data.discipline === "string") {
          setDetectedDiscipline(data.discipline as Discipline);
        }
        setDetectedConfidence(typeof data.confidence === "number" ? data.confidence : 0);
        setLowConfidence(Boolean(data.lowConfidence));
      } catch {
        // Non-blocking
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
      const payload: Record<string, unknown> = { resumeId, jobText };
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
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="private" />
        <main className="flex min-h-[calc(100vh-5rem)] items-center justify-center py-10">
          <div className="rt-panel w-full max-w-3xl p-5 sm:p-8">
            <p className="rt-label">New run</p>
            <h1 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)] sm:text-4xl">
              Paste job description
            </h1>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              We&apos;ll tailor your resume to match this job description.
            </p>

            <div className="mt-6">
              <label htmlFor="job-text" className="text-sm font-medium text-[var(--rt-ink-700)]">
                Job description
              </label>
              <textarea
                id="job-text"
                value={jobText}
                onChange={(e) => setJobText(e.target.value)}
                rows={10}
                className="mt-1 w-full rounded-xl border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2.5 text-[var(--rt-ink-900)] placeholder:text-[var(--rt-ink-500)] focus:border-[var(--rt-accent)] focus:outline-none focus:ring-1 focus:ring-[rgba(190,76,47,0.3)]"
                placeholder="Paste the job description here..."
              />
            </div>

            <div className="mt-4 rounded-xl border border-[rgba(67,63,52,0.18)] bg-[rgba(255,255,255,0.5)] p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-xs uppercase tracking-[0.18em] text-[var(--rt-ink-500)]">Detected discipline</p>
                {detecting ? <p className="text-xs text-[var(--rt-ink-500)]">Detecting&hellip;</p> : null}
              </div>
              <p className="mt-1 text-sm text-[var(--rt-ink-900)]">
                {detectedDiscipline ? labelForDiscipline(detectedDiscipline) : "Add more job details to detect"}
              </p>
              {detectedDiscipline ? (
                <p className={`mt-1 text-xs ${lowConfidence ? "text-amber-600" : "text-[var(--rt-ink-500)]"}`}>
                  Confidence: {Math.round(Math.max(0, Math.min(1, detectedConfidence)) * 100)}%
                  {lowConfidence ? " (low confidence)" : ""}
                </p>
              ) : null}

              <label htmlFor="discipline-override" className="mt-4 block text-xs font-medium uppercase tracking-[0.15em] text-[var(--rt-ink-500)]">
                Override discipline (optional)
              </label>
              <select
                id="discipline-override"
                value={disciplineOverride}
                onChange={(e) => setDisciplineOverride(e.target.value as "" | Discipline)}
                className="mt-1 w-full rounded-lg border border-[rgba(67,63,52,0.25)] bg-white/70 px-3 py-2 text-sm text-[var(--rt-ink-900)] focus:border-[var(--rt-accent)] focus:outline-none focus:ring-1 focus:ring-[rgba(190,76,47,0.3)]"
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
              <div role="alert" className="mt-4 rounded-lg border border-[rgba(190,76,47,0.2)] bg-[rgba(255,236,229,0.85)] px-3 py-2">
                <p className="text-sm text-[var(--rt-accent-strong)]">{error}</p>
              </div>
            ) : null}

            <div className="mt-6 flex justify-end">
              <button
                onClick={handleGenerate}
                disabled={loading || jobText.trim().length === 0}
                className="rt-btn-primary flex items-center gap-2 px-5 py-2.5 text-sm font-semibold disabled:opacity-70 focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--rt-accent)] focus-visible:ring-offset-2"
              >
                {loading ? (
                  <>
                    <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                    Generating&hellip;
                  </>
                ) : "Generate"}
              </button>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

export default function JobPage() {
  return (
    <Suspense fallback={<div className="rt-canvas"><div className="rt-shell" /></div>}>
      <JobPageInner />
    </Suspense>
  );
}
