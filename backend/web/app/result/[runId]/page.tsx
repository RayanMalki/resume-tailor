"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import TopBar from "../../components/TopBar";
import { useToast } from "../../components/Toast";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type TermScore = {
  term: string;
  score: number;
};

type BM25Signals = {
  top_job_terms: TermScore[];
  missing_job_terms: TermScore[];
  overlap_terms: string[];
  score: number;
};

type ATSReport = {
  score: number;
  notes: string[];
  change_plan: string[];
  summary: string;
  bm25_signals: BM25Signals | null;
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
  const { toast } = useToast();

  // Core data
  const [latex, setLatex] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState(5);
  const [progressStage, setProgressStage] = useState("Queued");
  const [resumeId, setResumeId] = useState<string | null>(null);

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
    return progressStage;
  }, [latex, progressStage]);

  // Target progress ref — set by polling, animated smoothly toward by the interval
  const targetProgressRef = useRef(5);

  /* ── Smooth progress animation — ticks toward the target set by polling ── */
  useEffect(() => {
    let timer: NodeJS.Timeout | undefined;
    if (!latex) {
      timer = setInterval(() => {
        setProgress((prev) => {
          const target = targetProgressRef.current;
          if (prev >= target) return prev;
          // Move 1-3% toward the target each tick for a smooth feel
          return Math.min(target, prev + Math.max(1, Math.floor((target - prev) / 4)));
        });
      }, 300);
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

        // The DB stores a reportPayload wrapper: { ats_report, bm25_signals, change_plan, ... }
        // Navigate into the nested ats_report for AI-generated score/notes.
        const inner = obj.ats_report ?? obj;
        const changePlanRaw = obj.change_plan;
        const changes: string[] = Array.isArray(changePlanRaw)
          ? changePlanRaw
          : Array.isArray(changePlanRaw?.changes)
            ? changePlanRaw.changes
            : [];

        // Extract BM25 signals
        const bm25Raw = obj.bm25_signals;
        const bm25: BM25Signals | null = bm25Raw && typeof bm25Raw === "object"
          ? {
              top_job_terms: Array.isArray(bm25Raw.top_job_terms) ? bm25Raw.top_job_terms : [],
              missing_job_terms: Array.isArray(bm25Raw.missing_job_terms) ? bm25Raw.missing_job_terms : [],
              overlap_terms: Array.isArray(bm25Raw.overlap_terms) ? bm25Raw.overlap_terms : [],
              score: typeof bm25Raw.score === "number" ? bm25Raw.score : 0,
            }
          : null;

        return {
          score: typeof inner.score === "number" ? inner.score : 0,
          notes: Array.isArray(inner.notes) ? inner.notes : [],
          change_plan: changes,
          summary: typeof inner.summary === "string" ? inner.summary : "",
          bm25_signals: bm25,
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
        }
      } catch {
        // ignore metadata failures
      }
    };

    const updateProgress = (pct: number, stage: string) => {
      targetProgressRef.current = pct;
      setProgressStage(stage);
    };

    const poll = async () => {
      // Stop after 5 minutes
      if (Date.now() - pollStartRef.current > 5 * 60 * 1000) {
        setError("Timed out waiting for result. Please refresh to try again.");
        return;
      }

      try {
        // 1. Check for LaTeX artifact (final deliverable)
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
          targetProgressRef.current = 100;
          setProgress(100);
          setProgressStage("Ready");
          pollErrorCountRef.current = 0;
          return;
        }

        if (res.status === 404) {
          // LaTeX not ready yet — check run status and intermediate artifacts
          const runRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}`, {
            credentials: "include",
          });
          if (runRes.ok) {
            const runData = await runRes.json();
            extractResumeId(runData);

            if (runData.status === "failed") {
              setError(runData.errorMessage || "Run failed");
              return;
            }

            // Derive progress from run status
            const status = runData.status || runData.Status;
            if (status === "queued") {
              updateProgress(10, "Queued \u2014 waiting for worker\u2026");
            } else if (status === "running") {
              // Pipeline: BM25 → Resume Spec/LaTeX → Report
              // Check if report exists (last step — means resume is done too)
              try {
                const reportRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}/report`, {
                  credentials: "include",
                });
                if (reportRes.ok) {
                  updateProgress(85, "Finalizing report\u2026");
                } else {
                  updateProgress(40, "Analyzing keywords & generating tailored resume\u2026");
                }
              } catch {
                updateProgress(25, "Processing\u2026");
              }
            }
          }

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

  /* ── Actions ────────────────────────────────────────────────── */
  const handleCopy = async () => {
    if (!latex) return;
    await navigator.clipboard.writeText(latex);
    toast("LaTeX copied to clipboard");
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
      toast("PDF downloaded");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to download PDF";
      setPdfError(msg);
      toast(msg, "error");
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
              <p className="mt-2 text-xs text-slate-500">{progressStage}</p>
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

                    {/* Change Summary */}
                    {atsReport.summary && (
                      <div className="rounded-xl border border-white/10 bg-ink-900/50 p-4">
                        <h4 className="text-sm font-semibold text-slate-200">What changed</h4>
                        <p className="mt-2 text-sm leading-relaxed text-slate-300">
                          {atsReport.summary}
                        </p>
                      </div>
                    )}

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

                    {/* BM25 Keyword Signals */}
                    {atsReport.bm25_signals && (
                      <div>
                        <h4 className="text-sm font-semibold text-slate-200">Keyword Signals (BM25)</h4>
                        <p className="mt-1 text-xs text-slate-400">
                          How well your resume keywords match the job description.
                        </p>

                        {/* Overlap terms */}
                        {atsReport.bm25_signals.overlap_terms.length > 0 && (
                          <div className="mt-3">
                            <p className="text-xs font-medium text-emerald-300">
                              Matched keywords ({atsReport.bm25_signals.overlap_terms.length})
                            </p>
                            <div className="mt-1.5 flex flex-wrap gap-1.5">
                              {atsReport.bm25_signals.overlap_terms.map((term) => (
                                <span
                                  key={term}
                                  className="rounded-full bg-emerald-500/15 px-2.5 py-1 text-xs text-emerald-300 border border-emerald-500/20"
                                >
                                  {term}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Missing terms */}
                        {atsReport.bm25_signals.missing_job_terms.length > 0 && (
                          <div className="mt-3">
                            <p className="text-xs font-medium text-rose-300">
                              Missing from resume ({atsReport.bm25_signals.missing_job_terms.length})
                            </p>
                            <div className="mt-1.5 flex flex-wrap gap-1.5">
                              {atsReport.bm25_signals.missing_job_terms.slice(0, 15).map((ts) => (
                                <span
                                  key={ts.term}
                                  className="rounded-full bg-rose-500/15 px-2.5 py-1 text-xs text-rose-300 border border-rose-500/20"
                                >
                                  {ts.term}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Top job terms with scores */}
                        {atsReport.bm25_signals.top_job_terms.length > 0 && (
                          <div className="mt-3">
                            <p className="text-xs font-medium text-slate-300">Top job terms by importance</p>
                            <div className="mt-1.5 space-y-1">
                              {atsReport.bm25_signals.top_job_terms.map((ts) => {
                                const isMatched = atsReport.bm25_signals!.overlap_terms.includes(ts.term);
                                return (
                                  <div key={ts.term} className="flex items-center gap-2">
                                    <span className={`text-xs w-24 truncate ${isMatched ? "text-emerald-300" : "text-slate-400"}`}>
                                      {ts.term}
                                    </span>
                                    <div className="flex-1 h-1.5 rounded-full bg-white/5 overflow-hidden">
                                      <div
                                        className={`h-full rounded-full transition-all ${isMatched ? "bg-emerald-500/60" : "bg-slate-500/40"}`}
                                        style={{ width: `${Math.min(100, (ts.score / (atsReport.bm25_signals!.top_job_terms[0]?.score || 1)) * 100)}%` }}
                                      />
                                    </div>
                                  </div>
                                );
                              })}
                            </div>
                          </div>
                        )}
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

            </>
          )}
        </div>
      </main>
    </div>
  );
}
