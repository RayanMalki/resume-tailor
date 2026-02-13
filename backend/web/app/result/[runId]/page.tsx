"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import TopBar from "../../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type ProjectMode = "pinned" | "auto" | "exclude";

type ProjectControl = {
  name: string;
  mode: ProjectMode;
};

type LLMProjectReason = {
  name: string;
  reason: string;
};

type ProjectOutcome = {
  name: string;
  mode: ProjectMode;
  reason: string;
};

type ATSReport = {
  score: number;
  notes: string[];
  change_plan: string[];
};

type Tab = "preview" | "report" | "latex";

/* ------------------------------------------------------------------ */
/*  Skeleton components for loading states                            */
/* ------------------------------------------------------------------ */

function SkeletonLine({ className = "" }: { className?: string }) {
  return (
    <div
      className={`animate-pulse rounded-lg bg-white/5 ${className}`}
    />
  );
}

function SkeletonBlock() {
  return (
    <div className="space-y-3">
      <SkeletonLine className="h-4 w-3/4" />
      <SkeletonLine className="h-4 w-full" />
      <SkeletonLine className="h-4 w-5/6" />
      <SkeletonLine className="h-4 w-2/3" />
    </div>
  );
}

function PDFSkeleton() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-2xl border border-white/10 bg-ink-950 p-10">
      <div className="h-10 w-10 animate-spin rounded-full border-2 border-ember-500/30 border-t-ember-500" />
      <p className="text-sm text-slate-400">Loading PDF preview&hellip;</p>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Score Ring                                                         */
/* ------------------------------------------------------------------ */

function ScoreRing({ score }: { score: number }) {
  const pct = Math.max(0, Math.min(100, Math.round(score * 100)));
  const radius = 40;
  const circumference = 2 * Math.PI * radius;
  const dashOffset = circumference - (pct / 100) * circumference;
  const color =
    pct >= 75
      ? "text-emerald-400"
      : pct >= 50
        ? "text-amber-400"
        : "text-rose-400";

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width="100" height="100" viewBox="0 0 100 100" role="img" aria-label={`ATS score: ${pct}%`}>
        <circle
          cx="50"
          cy="50"
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth="6"
          className="text-white/10"
        />
        <circle
          cx="50"
          cy="50"
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth="6"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={dashOffset}
          className={`${color} transition-all duration-700`}
          transform="rotate(-90 50 50)"
        />
      </svg>
      <span className={`absolute text-xl font-bold ${color}`}>{pct}%</span>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Main page                                                          */
/* ------------------------------------------------------------------ */

export default function ResultPage() {
  const router = useRouter();
  const params = useParams<{ runId: string }>();
  const runId = params.runId;

  // Core data
  const [latex, setLatex] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState(10);
  const [resumeId, setResumeId] = useState<string | null>(null);
  const [projectControls, setProjectControls] = useState<ProjectControl[]>([]);
  const [projectReasons, setProjectReasons] = useState<LLMProjectReason[]>([]);

  // PDF state
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const [pdfLoading, setPdfLoading] = useState(false);
  const [pdfError, setPdfError] = useState<string | null>(null);

  // ATS report state
  const [atsReport, setAtsReport] = useState<ATSReport | null>(null);
  const [reportLoading, setReportLoading] = useState(false);

  // Tab state
  const [activeTab, setActiveTab] = useState<Tab>("preview");

  // Polling ref for backoff
  const pollIntervalRef = useRef(1000);
  const pollErrorCountRef = useRef(0);
  const pollTimerRef = useRef<NodeJS.Timeout | null>(null);
  const pollStartRef = useRef(Date.now());

  // Cleanup PDF blob URL on unmount
  useEffect(() => {
    return () => {
      if (pdfUrl) URL.revokeObjectURL(pdfUrl);
    };
  }, [pdfUrl]);

  const loadingMessage = useMemo(() => {
    if (latex) return "Ready";
    return "Generating your tailored resume\u2026";
  }, [latex]);

  const projectOutcomes = useMemo<ProjectOutcome[]>(() => {
    if (projectReasons.length === 0) return [];
    const modeByName = new Map<string, ProjectMode>();
    for (const control of projectControls) {
      modeByName.set(control.name.toLowerCase(), control.mode);
    }
    return projectReasons
      .filter((item) => item.name.trim().length > 0 && item.reason.trim().length > 0)
      .map((item) => ({
        name: item.name,
        mode: modeByName.get(item.name.toLowerCase()) || "auto",
        reason: item.reason,
      }));
  }, [projectControls, projectReasons]);

  /* ── Progress bar animation ─────────────────────────────────── */
  useEffect(() => {
    let timer: NodeJS.Timeout | undefined;
    if (!latex) {
      timer = setInterval(() => {
        setProgress((prev) => (prev >= 90 ? prev : prev + 5));
      }, 500);
    }
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [latex]);

  /* ── Fetch PDF after LaTeX is ready ─────────────────────────── */
  const fetchPDF = useCallback(async () => {
    setPdfLoading(true);
    setPdfError(null);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-pdf`, {
        credentials: "include",
      });
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (res.status === 404) {
        // PDF not generated yet — retry after a moment
        setTimeout(fetchPDF, 3000);
        return;
      }
      if (!res.ok) throw new Error("Failed to fetch PDF");
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], { type: "application/pdf" });
      const url = URL.createObjectURL(blob);
      setPdfUrl(url);
    } catch (err) {
      setPdfError(err instanceof Error ? err.message : "Failed to load PDF");
    } finally {
      setPdfLoading(false);
    }
  }, [runId, router]);

  /* ── Fetch ATS report ───────────────────────────────────────── */
  const fetchReport = useCallback(async () => {
    setReportLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/report`, {
        credentials: "include",
      });
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (!res.ok) return;
      const data = await res.json();
      const raw = data.ATSReport ?? data.atsReport ?? data.ats_report;
      const parsed: ATSReport | null = (() => {
        if (!raw) return null;
        const obj = typeof raw === "string" ? JSON.parse(raw) : raw;
        if (!obj || typeof obj !== "object") return null;
        return {
          score: typeof obj.score === "number" ? obj.score : 0,
          notes: Array.isArray(obj.notes) ? obj.notes : [],
          change_plan: Array.isArray(obj.change_plan) ? obj.change_plan : [],
        };
      })();
      if (parsed) setAtsReport(parsed);
    } catch {
      // non-blocking
    } finally {
      setReportLoading(false);
    }
  }, [runId, router]);

  /* ── Main polling loop with exponential backoff (#12 fix) ──── */
  useEffect(() => {
    pollStartRef.current = Date.now();
    pollIntervalRef.current = 1000;
    pollErrorCountRef.current = 0;

    const extractResumeId = (data: Record<string, unknown>) => {
      const raw =
        (data.resumeId as string | undefined) ||
        (data.ResumeID as string | undefined) ||
        (data.resume_id as string | undefined);
      if (raw) setResumeId((prev) => prev || raw);
    };

    const extractProjectControls = (data: Record<string, unknown>) => {
      const raw = data.projectControls || data.ProjectControls || data.project_controls;
      if (!Array.isArray(raw)) return;
      const parsed: ProjectControl[] = raw
        .map((entry) => {
          if (!entry || typeof entry !== "object") return null;
          const obj = entry as Record<string, unknown>;
          const name = typeof obj.name === "string" ? obj.name.trim() : "";
          const modeRaw = typeof obj.mode === "string" ? obj.mode.trim().toLowerCase() : "auto";
          const mode: ProjectMode =
            modeRaw === "pinned" || modeRaw === "exclude" || modeRaw === "auto"
              ? (modeRaw as ProjectMode)
              : "auto";
          if (!name) return null;
          return { name, mode };
        })
        .filter((entry): entry is ProjectControl => Boolean(entry));
      setProjectControls(parsed);
    };

    const fetchRunMeta = async () => {
      try {
        const runRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}`, {
          credentials: "include",
        });
        if (runRes.status === 401) {
          router.push("/login");
          return;
        }
        if (runRes.ok) {
          const runData = await runRes.json();
          extractResumeId(runData);
          extractProjectControls(runData);
        }
      } catch {
        // ignore metadata failures
      }
    };

    const poll = async () => {
      // Stop after 5 minutes
      if (Date.now() - pollStartRef.current > 5 * 60 * 1000) {
        setError("Timed out waiting for result. Please refresh to try again.");
        return;
      }

      try {
        const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-latex`, {
          credentials: "include",
        });

        if (res.status === 401) {
          router.push("/login");
          return;
        }

        if (res.ok) {
          const data = await res.json();
          setLatex(data.latex);
          setProgress(100);
          pollErrorCountRef.current = 0;
          return;
        }

        if (res.status === 404) {
          const runRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}`, {
            credentials: "include",
          });
          if (runRes.ok) {
            const runData = await runRes.json();
            extractResumeId(runData);
            extractProjectControls(runData);
            if (runData.status === "failed") {
              setError(runData.errorMessage || "Run failed");
              return;
            }
          }
          // Schedule next poll with backoff
          scheduleNextPoll();
          return;
        }

        // Other error — increment error count
        pollErrorCountRef.current++;
        if (pollErrorCountRef.current >= 5) {
          setError("Failed to fetch result after multiple attempts.");
          return;
        }
        scheduleNextPoll();
      } catch (err) {
        pollErrorCountRef.current++;
        if (pollErrorCountRef.current >= 5) {
          setError(err instanceof Error ? err.message : "Failed to fetch artifact");
          return;
        }
        scheduleNextPoll();
      }
    };

    const scheduleNextPoll = () => {
      // Exponential backoff: 1s → 2s → 4s → 8s, capped at 10s
      pollIntervalRef.current = Math.min(pollIntervalRef.current * 2, 10000);
      pollTimerRef.current = setTimeout(poll, pollIntervalRef.current);
    };

    // Initial calls
    poll();
    fetchRunMeta();

    return () => {
      if (pollTimerRef.current) clearTimeout(pollTimerRef.current);
    };
  }, [runId, router]);

  /* ── Fetch PDF and report once LaTeX is ready ───────────────── */
  useEffect(() => {
    if (!latex) return;
    fetchPDF();
    fetchReport();
  }, [latex, fetchPDF, fetchReport]);

  /* ── Project reasons polling (with backoff) ─────────────────── */
  useEffect(() => {
    if (!latex) return;

    let canceled = false;
    let interval = 1500;
    let timer: NodeJS.Timeout;

    const fetchReasons = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/artifacts/project-reasons`, {
          credentials: "include",
        });
        if (res.status === 401) { router.push("/login"); return; }
        if (!res.ok) {
          interval = Math.min(interval * 2, 10000);
          if (!canceled) timer = setTimeout(fetchReasons, interval);
          return;
        }
        const data = (await res.json()) as { projects?: Array<{ name?: string; reason?: string }> };
        const parsed = (data.projects || [])
          .map((item) => ({
            name: (item.name || "").trim(),
            reason: (item.reason || "").trim(),
          }))
          .filter((item) => item.name.length > 0 && item.reason.length > 0);

        if (!canceled) setProjectReasons(parsed);
        if (parsed.length === 0 && !canceled) {
          interval = Math.min(interval * 2, 10000);
          timer = setTimeout(fetchReasons, interval);
        }
      } catch {
        if (!canceled) {
          interval = Math.min(interval * 2, 10000);
          timer = setTimeout(fetchReasons, interval);
        }
      }
    };

    fetchReasons();

    return () => {
      canceled = true;
      clearTimeout(timer);
    };
  }, [latex, runId, router]);

  /* ── Actions ────────────────────────────────────────────────── */
  const handleCopy = async () => {
    if (!latex) return;
    await navigator.clipboard.writeText(latex);
  };

  const handleDownloadPDF = async () => {
    if (pdfUrl) {
      const link = document.createElement("a");
      link.href = pdfUrl;
      link.download = "resume.pdf";
      document.body.appendChild(link);
      link.click();
      link.remove();
      return;
    }
    // Fallback: fetch fresh
    setPdfLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-pdf`, {
        credentials: "include",
      });
      if (!res.ok) throw new Error(res.status === 404 ? "PDF not ready yet." : "Failed to fetch PDF");
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], { type: "application/pdf" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = "resume.pdf";
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
    } catch (err) {
      setPdfError(err instanceof Error ? err.message : "Failed to download PDF");
    } finally {
      setPdfLoading(false);
    }
  };

  const handleGoToResume = () => router.push("/resume");
  const handleGoToJob = () => {
    if (resumeId) { router.push(`/job?resumeId=${resumeId}`); return; }
    router.push("/job");
  };

  /* ── Tabs ───────────────────────────────────────────────────── */
  const tabs: { id: Tab; label: string }[] = [
    { id: "preview", label: "PDF Preview" },
    { id: "report", label: "ATS Report" },
    { id: "latex", label: "LaTeX Source" },
  ];

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar showLogout />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-4 py-8 sm:px-6 sm:py-16">
        <div className="w-full max-w-3xl rounded-[28px] border border-white/10 bg-ink-900/70 p-5 shadow-panel backdrop-blur sm:p-8">

          {/* Header */}
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h1 className="text-xl font-semibold text-white sm:text-2xl">Your tailored resume</h1>
            <div className="flex flex-wrap items-center gap-2">
              <button
                onClick={handleGoToResume}
                className="rounded-full border border-white/10 px-3 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200 sm:px-4 sm:py-1.5 sm:tracking-[0.2em]"
                aria-label="Upload a new CV"
              >
                New CV
              </button>
              <button
                onClick={handleGoToJob}
                className="rounded-full border border-white/10 px-3 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-slate-200 transition hover:border-ember-400/60 hover:text-ember-200 sm:px-4 sm:py-1.5 sm:tracking-[0.2em]"
                aria-label="Upload a new job listing"
              >
                New Job
              </button>
            </div>
          </div>
          <p className="mt-2 text-sm text-slate-400">{loadingMessage}</p>

          {/* Loading state */}
          {!latex ? (
            <div className="mt-6">
              <div className="h-2 w-full overflow-hidden rounded-full bg-ink-950" role="progressbar" aria-valuenow={progress} aria-valuemin={0} aria-valuemax={100}>
                <div
                  className="h-full rounded-full bg-ember-500 transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
              {error ? (
                <div className="mt-4">
                  <p className="text-sm text-rose-300">{error}</p>
                  <button
                    onClick={() => window.location.reload()}
                    className="mt-2 rounded-full border border-white/10 px-4 py-1.5 text-xs font-semibold text-slate-200 transition hover:border-ember-400/60"
                  >
                    Retry
                  </button>
                </div>
              ) : (
                <div className="mt-6">
                  <SkeletonBlock />
                </div>
              )}
            </div>
          ) : (
            <>
              {/* Tab navigation */}
              <div className="mt-6 flex gap-1 rounded-xl border border-white/10 bg-ink-950/60 p-1" role="tablist" aria-label="Result tabs">
                {tabs.map((tab) => (
                  <button
                    key={tab.id}
                    role="tab"
                    aria-selected={activeTab === tab.id}
                    aria-controls={`panel-${tab.id}`}
                    onClick={() => setActiveTab(tab.id)}
                    className={`flex-1 rounded-lg px-3 py-2 text-xs font-semibold uppercase tracking-[0.15em] transition sm:text-sm ${
                      activeTab === tab.id
                        ? "bg-ember-500 text-ink-950 shadow-glow"
                        : "text-slate-400 hover:text-slate-200"
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              {/* Action buttons */}
              <div className="mt-4 flex flex-wrap items-center gap-2">
                <button
                  onClick={handleDownloadPDF}
                  disabled={pdfLoading}
                  className="rounded-full border border-ember-500/60 px-4 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-ember-200 transition hover:border-ember-400 hover:text-white disabled:opacity-50 sm:tracking-[0.2em]"
                  aria-label="Download resume as PDF"
                >
                  {pdfLoading ? "Preparing\u2026" : "Download PDF"}
                </button>
                <button
                  onClick={handleCopy}
                  className="rounded-full border border-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-slate-200 transition hover:border-ember-400 hover:text-ember-200 sm:tracking-[0.2em]"
                  aria-label="Copy LaTeX source to clipboard"
                >
                  Copy LaTeX
                </button>
              </div>
              {pdfError && <p className="mt-2 text-xs text-rose-300">{pdfError}</p>}

              {/* ── PDF Preview Tab ──────────────────────────── */}
              <div
                id="panel-preview"
                role="tabpanel"
                aria-labelledby="tab-preview"
                className={activeTab === "preview" ? "mt-4" : "hidden"}
              >
                {pdfLoading || !pdfUrl ? (
                  <PDFSkeleton />
                ) : (
                  <iframe
                    src={pdfUrl}
                    title="Resume PDF preview"
                    className="h-[600px] w-full rounded-2xl border border-white/10 bg-white sm:h-[750px]"
                  />
                )}
              </div>

              {/* ── ATS Report Tab ───────────────────────────── */}
              <div
                id="panel-report"
                role="tabpanel"
                aria-labelledby="tab-report"
                className={activeTab === "report" ? "mt-4" : "hidden"}
              >
                {reportLoading || !atsReport ? (
                  <div className="space-y-4 rounded-2xl border border-white/10 bg-ink-950/60 p-6">
                    <SkeletonBlock />
                  </div>
                ) : (
                  <div className="space-y-6 rounded-2xl border border-white/10 bg-ink-950/60 p-6">
                    {/* Score */}
                    <div className="flex flex-col items-center gap-3 sm:flex-row sm:gap-6">
                      <ScoreRing score={atsReport.score} />
                      <div>
                        <h3 className="text-lg font-semibold text-white">ATS Compatibility Score</h3>
                        <p className="mt-1 text-sm text-slate-400">
                          Based on BM25 keyword analysis and resume-to-job alignment.
                        </p>
                      </div>
                    </div>

                    {/* Notes */}
                    {atsReport.notes.length > 0 && (
                      <div>
                        <h4 className="text-sm font-semibold text-slate-200">Analysis Notes</h4>
                        <ul className="mt-2 space-y-2">
                          {atsReport.notes.map((note, i) => (
                            <li key={i} className="flex items-start gap-2 text-sm text-slate-300">
                              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-ember-400" />
                              {note}
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}

                    {/* Change plan */}
                    {atsReport.change_plan.length > 0 && (
                      <div>
                        <h4 className="text-sm font-semibold text-slate-200">Change Plan</h4>
                        <ul className="mt-2 space-y-2">
                          {atsReport.change_plan.map((change, i) => (
                            <li key={i} className="flex items-start gap-2 text-sm text-slate-300">
                              <span className="mt-1 flex-shrink-0 text-ember-400">&#10003;</span>
                              {change}
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
              </div>

              {/* ── LaTeX Source Tab ──────────────────────────── */}
              <div
                id="panel-latex"
                role="tabpanel"
                aria-labelledby="tab-latex"
                className={activeTab === "latex" ? "mt-4" : "hidden"}
              >
                <pre className="max-h-[500px] overflow-auto rounded-2xl border border-white/10 bg-ink-950 p-4 text-xs text-slate-100 sm:max-h-[600px]">
                  {latex}
                </pre>
              </div>

              {/* ── Project outcomes section ──────────────────── */}
              {projectOutcomes.length > 0 ? (
                <div className="mt-6 rounded-2xl border border-white/10 bg-ink-950/60 p-4">
                  <h2 className="text-sm font-semibold text-slate-200">Project decisions &amp; LLM reasoning</h2>
                  <p className="mt-1 text-xs text-slate-400">
                    Tailored explanations for each project based on your resume, job posting, and BM25 signals.
                  </p>
                  <div className="mt-3 space-y-2">
                    {projectOutcomes.map((outcome) => (
                      <div
                        key={outcome.name}
                        className="rounded-lg border border-white/10 bg-ink-900/70 px-3 py-2"
                      >
                        <div className="flex items-center justify-between gap-2">
                          <span className="text-sm text-slate-100">{outcome.name}</span>
                          <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-emerald-200">
                            {outcome.mode.toUpperCase()}
                          </span>
                        </div>
                        <p className="mt-1 text-xs text-slate-300">{outcome.reason}</p>
                      </div>
                    ))}
                  </div>
                </div>
              ) : projectControls.length > 0 ? (
                <div className="mt-6 rounded-2xl border border-white/10 bg-ink-950/60 p-4">
                  <h2 className="text-sm font-semibold text-slate-200">Project decisions &amp; LLM reasoning</h2>
                  <div className="mt-2 space-y-2">
                    <SkeletonLine className="h-3 w-2/3" />
                    <SkeletonLine className="h-3 w-1/2" />
                  </div>
                </div>
              ) : null}
            </>
          )}
        </div>
      </main>
    </div>
  );
}
