"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import EditorialNav from "../../components/EditorialNav";
import { useToast } from "../../components/Toast";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

type TermScore = {
  term: string;
  score: number;
};

type BM25Signals = {
  top_job_terms: TermScore[];
  missing_job_terms: TermScore[];
  overlap_terms: string[];
  bucketed_top_terms?: Record<string, TermScore[]>;
  low_signal_terms?: TermScore[];
  category_coverage?: Record<string, number>;
  discipline?: string;
  discipline_evidence?: TermScore[];
  profile_version?: string;
  score: number;
};

type InterviewQuestion = {
  question: string;
  answer_star: string[];
};

type ATSReport = {
  score: number;
  notes: string[];
  change_plan: string[];
  summary: string;
  interview_questions: InterviewQuestion[];
  bm25_signals: BM25Signals | null;
  discipline?: string;
  discipline_confidence?: number;
  discipline_source?: string;
  low_confidence?: boolean;
  category_coverage?: Record<string, number>;
  discipline_evidence?: TermScore[];
  profile_version?: string;
  scoring_discipline?: string;
};

type Tab = "preview" | "report" | "cover" | "latex";

const scoreHeadline = (score: number | null | undefined): string => {
  if (score == null) return "Analyzing your resume against the job";
  const pct = Math.round(score * 100);
  if (pct >= 91) return "Outstanding match — you're built for this role";
  if (pct >= 81) return "Excellent alignment — just polish the final details";
  if (pct >= 71) return "Strong base with room to sharpen impact language";
  if (pct >= 61) return "Good start — a few targeted edits will lift you higher";
  if (pct >= 51) return "Halfway there — key terms and structure need attention";
  if (pct >= 41) return "Decent foundation, but alignment gaps are holding you back";
  if (pct >= 31) return "Some signal detected — missing critical match factors";
  if (pct >= 21) return "Below the bar — gaps in core keywords and structure";
  if (pct >= 11) return "Far from the target — major restructuring required";
  return "Needs significant work — let's rebuild the alignment";
};

const bucketOrder: Array<{ key: string; label: string }> = [
  { key: "languages", label: "Languages" },
  { key: "cloud_devops_db", label: "Cloud / DevOps / DB" },
  { key: "practices", label: "Practices" },
  { key: "soft_skills", label: "Soft Skills" },
  { key: "other", label: "Other (high signal only)" },
];

const disciplineLabels: Record<string, string> = {
  mechanical: "Mechanical Engineering",
  electrical: "Electrical Engineering",
  industrial_logistics: "Industrial / Operations / Logistics",
  aerospace: "Aerospace Engineering",
  it_software: "IT / Software Engineering",
};

/* ------------------------------------------------------------------ */
/*  Skeleton components for loading states                            */
/* ------------------------------------------------------------------ */

function SkeletonLine({ className = "" }: { className?: string }) {
  return (
    <div
      className={`animate-pulse rounded-lg bg-[rgba(81,73,62,0.12)] ${className}`}
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
    <div className="flex flex-col items-center justify-center gap-4 rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.58)] p-10">
      <div className="h-10 w-10 animate-spin rounded-full border-2 border-[rgba(46,102,210,0.2)] border-t-[var(--rt-blue)]" />
      <p className="text-sm text-[var(--rt-ink-700)]">
        Loading resume preview&hellip;
      </p>
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
      ? "text-[var(--rt-green)]"
      : pct >= 50
        ? "text-[var(--rt-blue)]"
        : "text-[var(--rt-accent-strong)]";

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg
        width="100"
        height="100"
        viewBox="0 0 100 100"
        role="img"
        aria-label={`ATS score: ${pct}%`}
      >
        <circle
          cx="50"
          cy="50"
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth="6"
          className="rt-score-track"
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
  const [coverLetterPdfUrl, setCoverLetterPdfUrl] = useState<string | null>(
    null,
  );
  const [coverLetterPdfLoading, setCoverLetterPdfLoading] = useState(false);

  // Tab state
  const [activeTab, setActiveTab] = useState<Tab>("preview");

  // Polling ref for backoff
  const pollIntervalRef = useRef(1000);
  const pollErrorCountRef = useRef(0);
  const pollTimerRef = useRef<NodeJS.Timeout | null>(null);
  const pollStartRef = useRef(Date.now());

  // Cleanup PDF blob URLs on unmount
  useEffect(() => {
    return () => {
      if (pdfUrl) URL.revokeObjectURL(pdfUrl);
    };
  }, [pdfUrl]);

  useEffect(() => {
    return () => {
      if (coverLetterPdfUrl) URL.revokeObjectURL(coverLetterPdfUrl);
    };
  }, [coverLetterPdfUrl]);

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
          return Math.min(
            target,
            prev + Math.max(1, Math.floor((target - prev) / 4)),
          );
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
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-pdf`,
        {
          credentials: "include",
        },
      );
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
        const bm25: BM25Signals | null =
          bm25Raw && typeof bm25Raw === "object"
            ? {
                top_job_terms: Array.isArray(bm25Raw.top_job_terms)
                  ? bm25Raw.top_job_terms
                  : [],
                missing_job_terms: Array.isArray(bm25Raw.missing_job_terms)
                  ? bm25Raw.missing_job_terms
                  : [],
                overlap_terms: Array.isArray(bm25Raw.overlap_terms)
                  ? bm25Raw.overlap_terms
                  : [],
                bucketed_top_terms:
                  bm25Raw.bucketed_top_terms &&
                  typeof bm25Raw.bucketed_top_terms === "object"
                    ? (bm25Raw.bucketed_top_terms as Record<
                        string,
                        TermScore[]
                      >)
                    : undefined,
                low_signal_terms: Array.isArray(bm25Raw.low_signal_terms)
                  ? bm25Raw.low_signal_terms
                  : [],
                category_coverage:
                  bm25Raw.category_coverage &&
                  typeof bm25Raw.category_coverage === "object"
                    ? (bm25Raw.category_coverage as Record<string, number>)
                    : undefined,
                discipline:
                  typeof bm25Raw.discipline === "string"
                    ? bm25Raw.discipline
                    : undefined,
                discipline_evidence: Array.isArray(bm25Raw.discipline_evidence)
                  ? bm25Raw.discipline_evidence
                  : [],
                profile_version:
                  typeof bm25Raw.profile_version === "string"
                    ? bm25Raw.profile_version
                    : undefined,
                score: typeof bm25Raw.score === "number" ? bm25Raw.score : 0,
              }
            : null;

        const interviewQuestions = Array.isArray(inner.interview_questions)
          ? inner.interview_questions
              .filter((item: unknown) => item && typeof item === "object")
              .map((item: unknown) => {
                const asRecord = item as Record<string, unknown>;
                return {
                  question:
                    typeof asRecord.question === "string"
                      ? asRecord.question
                      : "",
                  answer_star: Array.isArray(asRecord.answer_star)
                    ? asRecord.answer_star.filter(
                        (entry): entry is string => typeof entry === "string",
                      )
                    : [],
                } as InterviewQuestion;
              })
              .filter((item: InterviewQuestion) => item.question.trim() !== "")
          : [];

        return {
          score: typeof inner.score === "number" ? inner.score : 0,
          notes: Array.isArray(inner.notes) ? inner.notes : [],
          change_plan: changes,
          summary: typeof inner.summary === "string" ? inner.summary : "",
          interview_questions: interviewQuestions,
          bm25_signals: bm25,
          discipline:
            typeof obj.discipline === "string" ? obj.discipline : undefined,
          discipline_confidence:
            typeof obj.discipline_confidence === "number"
              ? obj.discipline_confidence
              : undefined,
          discipline_source:
            typeof obj.discipline_source === "string"
              ? obj.discipline_source
              : undefined,
          low_confidence: Boolean(obj.low_confidence),
          category_coverage:
            obj.category_coverage && typeof obj.category_coverage === "object"
              ? (obj.category_coverage as Record<string, number>)
              : bm25?.category_coverage,
          discipline_evidence: Array.isArray(obj.discipline_evidence)
            ? obj.discipline_evidence
            : bm25?.discipline_evidence,
          profile_version:
            typeof obj.profile_version === "string"
              ? obj.profile_version
              : bm25?.profile_version,
          scoring_discipline:
            typeof obj.scoring_discipline === "string"
              ? obj.scoring_discipline
              : undefined,
        };
      })();
      if (parsed) setAtsReport(parsed);
    } catch {
      // non-blocking
    } finally {
      setReportLoading(false);
    }
  }, [runId, router]);

  const fetchCoverLetterPdf = useCallback(async () => {
    setCoverLetterPdfLoading(true);
    try {
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/cover-letter-pdf`,
        {
          credentials: "include",
        },
      );
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (res.status === 404) {
        // PDF not generated yet — retry after a moment
        setTimeout(fetchCoverLetterPdf, 3000);
        return;
      }
      if (!res.ok) return;
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], { type: "application/pdf" });
      const url = URL.createObjectURL(blob);
      setCoverLetterPdfUrl(url);
    } catch {
      // non-blocking
    } finally {
      setCoverLetterPdfLoading(false);
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
        const res = await fetch(
          `${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-latex`,
          {
            credentials: "include",
          },
        );

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
                const reportRes = await fetch(
                  `${API_BASE_URL}/v1/runs/${runId}/report`,
                  {
                    credentials: "include",
                  },
                );
                if (reportRes.ok) {
                  updateProgress(85, "Finalizing report\u2026");
                } else {
                  updateProgress(
                    40,
                    "Analyzing keywords & generating tailored resume\u2026",
                  );
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
          setError(
            err instanceof Error ? err.message : "Failed to fetch artifact",
          );
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
    fetchCoverLetterPdf();
  }, [latex, fetchPDF, fetchReport, fetchCoverLetterPdf]);

  /* ── Poll for ATS report until available (report is generated after LaTeX) ── */
  useEffect(() => {
    if (!latex || atsReport) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    let attempts = 0;
    const maxAttempts = 30; // ~60s total

    const pollReport = async () => {
      if (cancelled || attempts >= maxAttempts) return;
      attempts++;
      try {
        const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/report`, {
          credentials: "include",
        });
        if (res.status === 401) {
          router.push("/login");
          return;
        }
        if (res.ok) {
          await fetchReport();
          return;
        }
      } catch {
        // ignore, retry
      }
      if (!cancelled) {
        timer = setTimeout(pollReport, 2000);
      }
    };

    timer = setTimeout(pollReport, 2000);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [latex, atsReport, runId, router, fetchReport]);

  /* ── Poll for cover letter PDF until available (generated after LaTeX) ── */
  useEffect(() => {
    if (!latex || coverLetterPdfUrl) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    let attempts = 0;
    const maxAttempts = 30; // ~60s total

    const pollCoverLetterPdf = async () => {
      if (cancelled || attempts >= maxAttempts) return;
      attempts++;
      try {
        const res = await fetch(
          `${API_BASE_URL}/v1/runs/${runId}/artifacts/cover-letter-pdf`,
          {
            credentials: "include",
          },
        );
        if (res.status === 401) {
          router.push("/login");
          return;
        }
        if (res.ok) {
          await fetchCoverLetterPdf();
          return;
        }
      } catch {
        // ignore, retry
      }
      if (!cancelled) {
        timer = setTimeout(pollCoverLetterPdf, 2000);
      }
    };

    timer = setTimeout(pollCoverLetterPdf, 2000);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [latex, coverLetterPdfUrl, runId, router, fetchCoverLetterPdf]);

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
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-pdf`,
        {
          credentials: "include",
        },
      );
      if (!res.ok)
        throw new Error(
          res.status === 404 ? "PDF not ready yet." : "Failed to fetch PDF",
        );
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

  const handleDownloadDOCX = async () => {
    try {
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-docx`,
        {
          credentials: "include",
        },
      );
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (!res.ok)
        throw new Error(
          res.status === 404 ? "DOCX not ready yet." : "Failed to fetch DOCX",
        );
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], {
        type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = "resume.docx";
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      toast("DOCX downloaded");
    } catch (err) {
      toast(
        err instanceof Error ? err.message : "Failed to download DOCX",
        "error",
      );
    }
  };

  const handleDownloadCoverLetterPDF = async () => {
    try {
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/cover-letter-pdf`,
        {
          credentials: "include",
        },
      );
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (!res.ok)
        throw new Error(
          res.status === 404
            ? "Cover letter PDF not ready yet."
            : "Failed to fetch cover letter PDF",
        );
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], { type: "application/pdf" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = "cover_letter.pdf";
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      toast("Cover letter PDF downloaded");
    } catch (err) {
      toast(
        err instanceof Error
          ? err.message
          : "Failed to download cover letter PDF",
        "error",
      );
    }
  };

  const handleDownloadCoverLetterDOCX = async () => {
    try {
      const res = await fetch(
        `${API_BASE_URL}/v1/runs/${runId}/artifacts/cover-letter-docx`,
        {
          credentials: "include",
        },
      );
      if (res.status === 401) {
        router.push("/login");
        return;
      }
      if (!res.ok)
        throw new Error(
          res.status === 404
            ? "Cover letter DOCX not ready yet."
            : "Failed to fetch cover letter DOCX",
        );
      const buffer = await res.arrayBuffer();
      const blob = new Blob([buffer], {
        type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = "cover_letter.docx";
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
      toast("Cover letter DOCX downloaded");
    } catch (err) {
      toast(
        err instanceof Error
          ? err.message
          : "Failed to download cover letter DOCX",
        "error",
      );
    }
  };

  const handleGoToResume = () => router.push("/resume");
  const handleGoToJob = () => {
    if (resumeId) {
      router.push(`/job?resumeId=${resumeId}`);
      return;
    }
    router.push("/job");
  };

  /* ── Tabs ───────────────────────────────────────────────────── */
  const tabs: { id: Tab; label: string }[] = [
    { id: "preview", label: "Resume Preview" },
    { id: "report", label: "ATS Report" },
    { id: "cover", label: "Cover Letter Preview" },
    { id: "latex", label: "LaTeX Source" },
  ];

  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="private" />
        <main className="mt-4 space-y-4 sm:mt-6 sm:space-y-6">
          <section className="rt-panel rt-fade-up p-5 sm:p-7">
            <div className="flex flex-wrap items-end justify-between gap-3">
              <div>
                <p className="rt-label">ATS Snapshot</p>
                <h1 className="mt-2 font-grotesk text-4xl font-semibold text-[var(--rt-ink-900)] sm:text-5xl">
                  {scoreHeadline(atsReport?.score)}
                </h1>
                <p className="font-serif-display mt-2 text-xl text-[var(--rt-ink-700)]">
                  {loadingMessage}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <button
                  onClick={handleGoToResume}
                  className="rt-btn-secondary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                  aria-label="Upload a new resume"
                >
                  New resume
                </button>
                <button
                  onClick={handleGoToJob}
                  className="rt-btn-secondary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                  aria-label="Upload a new job listing"
                >
                  New job
                </button>
              </div>
            </div>

            {!latex ? (
              <div className="mt-6 rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.58)] p-4 sm:p-5">
                <div
                  className="rt-progress-track"
                  role="progressbar"
                  aria-valuenow={progress}
                  aria-valuemin={0}
                  aria-valuemax={100}
                >
                  <div
                    className="rt-progress-fill"
                    style={{ width: `${progress}%` }}
                  />
                </div>
                <p className="mt-2 text-sm text-[var(--rt-ink-700)]">
                  {progressStage}
                </p>
                {error ? (
                  <div className="mt-3">
                    <p className="text-sm text-[var(--rt-accent-strong)]">
                      {error}
                    </p>
                    <button
                      onClick={() => window.location.reload()}
                      className="rt-btn-secondary mt-2 px-4 py-2 text-xs font-semibold uppercase tracking-[0.16em]"
                    >
                      Retry
                    </button>
                  </div>
                ) : (
                  <div className="mt-4">
                    <SkeletonBlock />
                  </div>
                )}
              </div>
            ) : (
              <>
                <div
                  className="mt-6 flex flex-wrap gap-2 rounded-[1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.52)] p-1"
                  role="tablist"
                  aria-label="Result tabs"
                >
                  {tabs.map((tab) => (
                    <button
                      key={tab.id}
                      id={`tab-${tab.id}`}
                      role="tab"
                      aria-selected={activeTab === tab.id}
                      aria-controls={`panel-${tab.id}`}
                      onClick={() => setActiveTab(tab.id)}
                      className={`flex-1 rounded-[0.8rem] px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] transition sm:text-sm ${
                        activeTab === tab.id
                          ? "bg-[var(--rt-accent)] text-[#fff8f3]"
                          : "text-[var(--rt-ink-500)] hover:text-[var(--rt-ink-900)]"
                      }`}
                    >
                      {tab.label}
                    </button>
                  ))}
                </div>

                {activeTab === "preview" ? (
                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    <button
                      onClick={handleDownloadPDF}
                      disabled={pdfLoading}
                      className="rt-btn-primary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] disabled:cursor-not-allowed disabled:opacity-65"
                      aria-label="Download resume as PDF"
                    >
                      {pdfLoading ? "Preparing..." : "Download PDF"}
                    </button>
                    <button
                      onClick={handleDownloadDOCX}
                      className="rt-btn-secondary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                      aria-label="Download resume as DOCX"
                    >
                      Download DOCX
                    </button>
                    <button
                      onClick={handleCopy}
                      className="rt-btn-secondary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                      aria-label="Copy LaTeX source to clipboard"
                    >
                      Copy LaTeX
                    </button>
                  </div>
                ) : null}

                {pdfError && activeTab === "preview" ? (
                  <p className="mt-2 text-xs text-[var(--rt-accent-strong)]">
                    {pdfError}
                  </p>
                ) : null}

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
                      title="Resume preview"
                      className="h-[65vh] min-h-[520px] w-full rounded-[1.2rem] border border-[var(--rt-stroke)] bg-white"
                    />
                  )}
                </div>

                <div
                  id="panel-report"
                  role="tabpanel"
                  aria-labelledby="tab-report"
                  className={activeTab === "report" ? "mt-4" : "hidden"}
                >
                  {reportLoading || !atsReport ? (
                    <div className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.58)] p-5">
                      <SkeletonBlock />
                    </div>
                  ) : (
                    <div className="space-y-4">
                      <section className="grid gap-4 lg:grid-cols-[1.15fr_1fr]">
                        <article className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.66)] p-5">
                          <p className="rt-label">ATS Score</p>
                          <div className="mt-4 flex flex-col items-center gap-4 sm:flex-row sm:items-center">
                            <ScoreRing score={atsReport.score} />
                            <div>
                              <h3 className="font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
                                Match quality overview
                              </h3>
                            </div>
                          </div>
                        </article>

                        <article className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.66)] p-5">
                          <p className="rt-label">Coverage by bucket</p>
                          {atsReport.category_coverage &&
                          Object.keys(atsReport.category_coverage).length > 0 ? (
                            <div className="mt-4 space-y-3">
                              {Object.entries(atsReport.category_coverage)
                                .sort(([a], [b]) => a.localeCompare(b))
                                .map(([bucket, value]) => {
                                  const pct = Math.round(
                                    Math.max(0, Math.min(1, value)) * 100,
                                  );
                                  return (
                                    <div key={`coverage-${bucket}`}>
                                      <div className="mb-1 flex items-center justify-between gap-2 text-sm">
                                        <span className="font-grotesk text-xl font-semibold text-[var(--rt-ink-900)]">
                                          {bucketOrder.find(
                                            (item) => item.key === bucket,
                                          )?.label || bucket}
                                        </span>
                                        <span className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--rt-ink-500)]">
                                          {pct}%
                                        </span>
                                      </div>
                                      <div className="rt-progress-track">
                                        <div
                                          className="rt-progress-fill"
                                          style={{ width: `${pct}%` }}
                                        />
                                      </div>
                                    </div>
                                  );
                                })}
                            </div>
                          ) : (
                            <p className="mt-3 text-sm text-[var(--rt-ink-700)]">
                              Coverage data will appear once scoring buckets are available.
                            </p>
                          )}
                        </article>
                      </section>

                      {(atsReport.discipline || atsReport.profile_version) && (
                        <section className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] p-5">
                          <p className="rt-label">Discipline Context</p>
                          <p className="mt-2 font-grotesk text-2xl font-semibold text-[var(--rt-ink-900)]">
                            {atsReport.discipline
                              ? disciplineLabels[atsReport.discipline] ||
                                atsReport.discipline
                              : "Auto-detected profile"}
                          </p>
                          <p className="mt-2 text-sm text-[var(--rt-ink-700)]">
                            Confidence:{" "}
                            {Math.round(
                              (atsReport.discipline_confidence ?? 0) * 100,
                            )}
                            % • Source: {atsReport.discipline_source || "auto"}
                            {atsReport.low_confidence ? " • low confidence" : ""}
                            {atsReport.scoring_discipline &&
                            atsReport.scoring_discipline !== atsReport.discipline
                              ? ` • scoring profile: ${disciplineLabels[atsReport.scoring_discipline] || atsReport.scoring_discipline}`
                              : ""}
                          </p>
                          {atsReport.profile_version ? (
                            <p className="mt-1 text-xs text-[var(--rt-ink-500)]">
                              Profile version: {atsReport.profile_version}
                            </p>
                          ) : null}
                          {Array.isArray(atsReport.discipline_evidence) &&
                          atsReport.discipline_evidence.length > 0 ? (
                            <div className="mt-3 flex flex-wrap gap-2">
                              {atsReport.discipline_evidence
                                .slice(0, 8)
                                .map((item) => (
                                  <span
                                    key={`discipline-evidence-${item.term}`}
                                    className="rounded-full border border-[rgba(50,97,187,0.24)] bg-[rgba(216,228,247,0.72)] px-2.5 py-1 text-xs text-[rgba(34,74,148,0.92)]"
                                  >
                                    {item.term}
                                  </span>
                                ))}
                            </div>
                          ) : null}
                        </section>
                      )}

                      <section className="grid gap-4 lg:grid-cols-2">
                        {atsReport.summary ? (
                          <article className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] p-5">
                            <p className="rt-label">What changed</p>
                            <p className="font-serif-display mt-3 text-lg leading-relaxed text-[var(--rt-ink-700)]">
                              {atsReport.summary}
                            </p>
                          </article>
                        ) : null}

                        {atsReport.change_plan.length > 0 ? (
                          <article className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] p-5">
                            <p className="rt-label">Change plan</p>
                            <ul className="mt-3 space-y-2">
                              {atsReport.change_plan.map((change, index) => (
                                <li
                                  key={`plan-${index}`}
                                  className="flex items-start gap-2 text-sm text-[var(--rt-ink-700)]"
                                >
                                  <span className="mt-[0.35rem] inline-block h-1.5 w-1.5 rounded-full bg-[var(--rt-accent)]" />
                                  {change}
                                </li>
                              ))}
                            </ul>
                          </article>
                        ) : null}
                      </section>

                      {atsReport.notes.length > 0 ? (
                        <section className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] p-5">
                          <p className="rt-label">Analysis notes</p>
                          <ul className="mt-3 space-y-2">
                            {atsReport.notes.map((note, index) => (
                              <li
                                key={`note-${index}`}
                                className="flex items-start gap-2 text-sm text-[var(--rt-ink-700)]"
                              >
                                <span className="mt-[0.35rem] inline-block h-1.5 w-1.5 rounded-full bg-[var(--rt-blue)]" />
                                {note}
                              </li>
                            ))}
                          </ul>
                        </section>
                      ) : null}


                      {atsReport.bm25_signals ? (
                        <section className="rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] p-5">
                          <p className="rt-label">Keyword signal map</p>
                          <p className="mt-2 text-sm text-[var(--rt-ink-700)]">
                            BM25 terms show which language is already strong and which terms still need coverage.
                          </p>

                          {atsReport.bm25_signals.overlap_terms.length > 0 ? (
                            <div className="mt-4">
                              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--rt-green)]">
                                Matched keywords ({atsReport.bm25_signals.overlap_terms.length})
                              </p>
                              <div className="mt-2 flex flex-wrap gap-1.5">
                                {atsReport.bm25_signals.overlap_terms.map((term) => (
                                  <span
                                    key={term}
                                    className="rounded-full border border-[rgba(41,137,109,0.34)] bg-[rgba(220,243,234,0.76)] px-2.5 py-1 text-xs text-[rgba(28,112,87,0.92)]"
                                  >
                                    {term}
                                  </span>
                                ))}
                              </div>
                            </div>
                          ) : null}

                          {atsReport.bm25_signals.missing_job_terms.length > 0 ? (
                            <div className="mt-4">
                              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--rt-accent-strong)]">
                                Missing from resume ({atsReport.bm25_signals.missing_job_terms.length})
                              </p>
                              <div className="mt-2 flex flex-wrap gap-1.5">
                                {atsReport.bm25_signals.missing_job_terms
                                  .slice(0, 18)
                                  .map((termScore) => (
                                    <span
                                      key={termScore.term}
                                      className="rounded-full border border-[rgba(183,83,60,0.28)] bg-[rgba(255,237,232,0.82)] px-2.5 py-1 text-xs text-[var(--rt-accent-strong)]"
                                    >
                                      {termScore.term}
                                    </span>
                                  ))}
                              </div>
                            </div>
                          ) : null}

                          {atsReport.bm25_signals.top_job_terms.length > 0 ? (
                            <div className="mt-4">
                              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--rt-ink-500)]">
                                Top job terms by weight
                              </p>
                              <div className="mt-2 space-y-2">
                                {atsReport.bm25_signals.top_job_terms
                                  .slice(0, 10)
                                  .map((termScore) => {
                                    const topScore =
                                      atsReport.bm25_signals?.top_job_terms[0]
                                        ?.score || 1;
                                    const width = Math.min(
                                      100,
                                      (termScore.score / topScore) * 100,
                                    );
                                    const matched =
                                      atsReport.bm25_signals?.overlap_terms.includes(
                                        termScore.term,
                                      );

                                    return (
                                      <div
                                        key={`top-term-${termScore.term}`}
                                        className="grid grid-cols-[minmax(0,120px)_1fr] items-center gap-2"
                                      >
                                        <span
                                          className={`truncate text-xs ${
                                            matched
                                              ? "text-[var(--rt-green)]"
                                              : "text-[var(--rt-ink-700)]"
                                          }`}
                                        >
                                          {termScore.term}
                                        </span>
                                        <div className="rt-progress-track">
                                          <div
                                            className="rt-progress-fill"
                                            style={{ width: `${width}%` }}
                                          />
                                        </div>
                                      </div>
                                    );
                                  })}
                              </div>
                            </div>
                          ) : null}

                          {atsReport.bm25_signals.bucketed_top_terms ? (
                            <div className="mt-4 space-y-2">
                              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--rt-ink-500)]">
                                Deterministic skill buckets
                              </p>
                              {bucketOrder.map((bucket) => {
                                const items =
                                  atsReport.bm25_signals?.bucketed_top_terms?.[
                                    bucket.key
                                  ] || [];

                                if (items.length === 0) return null;

                                return (
                                  <div key={`bucket-${bucket.key}`}>
                                    <p className="text-[11px] uppercase tracking-[0.18em] text-[var(--rt-ink-500)]">
                                      {bucket.label}
                                    </p>
                                    <div className="mt-1 flex flex-wrap gap-1.5">
                                      {items.slice(0, 10).map((item) => {
                                        const matched =
                                          atsReport.bm25_signals?.overlap_terms.includes(
                                            item.term,
                                          );
                                        return (
                                          <span
                                            key={`${bucket.key}:${item.term}`}
                                            className={`rounded-full border px-2.5 py-1 text-xs ${
                                              matched
                                                ? "border-[rgba(41,137,109,0.34)] bg-[rgba(220,243,234,0.76)] text-[rgba(28,112,87,0.92)]"
                                                : "border-[rgba(67,63,52,0.2)] bg-[rgba(255,255,255,0.58)] text-[var(--rt-ink-700)]"
                                            }`}
                                          >
                                            {item.term}
                                          </span>
                                        );
                                      })}
                                    </div>
                                  </div>
                                );
                              })}
                            </div>
                          ) : null}

                          {Array.isArray(atsReport.bm25_signals.low_signal_terms) &&
                          atsReport.bm25_signals.low_signal_terms.length > 0 ? (
                            <p className="mt-4 text-xs text-[var(--rt-ink-500)]">
                              {atsReport.bm25_signals.low_signal_terms.length} low-signal terms were filtered from the primary term lists.
                            </p>
                          ) : null}
                        </section>
                      ) : null}
                    </div>
                  )}
                </div>

                <div
                  id="panel-cover"
                  role="tabpanel"
                  aria-labelledby="tab-cover"
                  className={activeTab === "cover" ? "mt-4" : "hidden"}
                >
                  {coverLetterPdfLoading || !coverLetterPdfUrl ? (
                    <PDFSkeleton />
                  ) : (
                    <div className="space-y-3">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <p className="rt-label">Generated cover letter</p>
                        <div className="flex flex-wrap gap-2">
                          <button
                            onClick={handleDownloadCoverLetterPDF}
                            className="rt-btn-primary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                            aria-label="Download cover letter as PDF"
                          >
                            Download PDF
                          </button>
                          <button
                            onClick={handleDownloadCoverLetterDOCX}
                            className="rt-btn-secondary px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em]"
                            aria-label="Download cover letter as DOCX"
                          >
                            Download DOCX
                          </button>
                        </div>
                      </div>
                      <iframe
                        src={coverLetterPdfUrl}
                        title="Cover letter preview"
                        className="h-[65vh] min-h-[520px] w-full rounded-[1.2rem] border border-[var(--rt-stroke)] bg-white"
                      />
                    </div>
                  )}
                </div>

                <div
                  id="panel-latex"
                  role="tabpanel"
                  aria-labelledby="tab-latex"
                  className={activeTab === "latex" ? "mt-4" : "hidden"}
                >
                  <pre className="max-h-[68vh] overflow-auto rounded-[1.2rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.68)] p-4 text-xs leading-relaxed text-[var(--rt-ink-900)]">
                    {latex}
                  </pre>
                </div>
              </>
            )}
          </section>
        </main>
      </div>
    </div>
  );
}
