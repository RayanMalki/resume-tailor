"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import EditorialNav from "../components/EditorialNav";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

type RunItem = {
  ID?: string; id?: string;
  ResumeID?: string; resumeId?: string;
  JobTitle?: string; jobTitle?: string;
  Company?: string; company?: string;
  Status?: string; status?: string;
  CreatedAt?: string; createdAt?: string;
  JobText?: string; jobText?: string;
};

const runTitle = (r: RunItem): string => {
  const explicit = r.JobTitle ?? r.jobTitle;
  if (explicit) return explicit;
  const desc = r.JobText ?? r.jobText ?? "";
  if (!desc) return "";
  // First sentence, capped at 6 words, then ellipsis
  const sentence = desc.split(/[.\n]/)[0].trim();
  const words = sentence.split(/\s+/).filter(Boolean);
  if (words.length <= 6) return words.join(" ");
  return words.slice(0, 6).join(" ") + "…";
};

const normalizeRun = (r: RunItem) => ({
  id: r.ID ?? r.id ?? "",
  jobTitle: runTitle(r),
  company: r.Company ?? r.company ?? "",
  status: r.Status ?? r.status ?? "",
  createdAt: r.CreatedAt ?? r.createdAt ?? "",
});

const formatRunDate = (iso: string): string => {
  if (!iso) return "";
  const date = new Date(iso);
  const now = new Date();
  const diffDays = Math.floor((now.setHours(0,0,0,0) - new Date(date).setHours(0,0,0,0)) / 86400000);
  const time = new Intl.DateTimeFormat("en-US", { hour: "numeric", minute: "2-digit" }).format(date);
  if (diffDays === 0) return `Today at ${time}`;
  if (diffDays === 1) return `Yesterday at ${time}`;
  const dateStr = new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(date);
  return `${dateStr} at ${time}`;
};

const statusBadge = (status: string) => {
  const s = status.toLowerCase();
  if (s === "succeeded" || s === "success") {
    return "bg-green-100 text-green-800 border border-green-200";
  }
  if (s === "needs_edits" || s === "failed") {
    return "bg-amber-100 text-amber-800 border border-amber-200";
  }
  return "bg-blue-100 text-blue-800 border border-blue-200";
};

type ResumeItem = {
  ID?: string;
  id?: string;
  Title?: string;
  title?: string;
};

type MeData = {
  onboardingSeen: boolean;
  hasApiKey: boolean;
  dailyRunsUsed: number;
  dailyRunsLimit: number;
};

type ScorePoint = {
  runId: string;
  score: number;
};

const parseScore = (atsRaw: unknown): number | null => {
  const parsed =
    typeof atsRaw === "string"
      ? (() => {
          try {
            return JSON.parse(atsRaw);
          } catch {
            return null;
          }
        })()
      : atsRaw;

  if (!parsed || typeof parsed !== "object") return null;

  const asRecord = parsed as Record<string, unknown>;
  const directScore = asRecord.score;
  if (typeof directScore === "number") return directScore;

  const nested = asRecord.ats_report || asRecord.atsReport || asRecord.ATSReport;
  if (nested && typeof nested === "object") {
    const nestedScore = (nested as Record<string, unknown>).score;
    if (typeof nestedScore === "number") return nestedScore;
  }

  return null;
};

export default function DashboardPage() {
  const router = useRouter();

  const [runs, setRuns] = useState<RunItem[]>([]);
  const [resumes, setResumes] = useState<ResumeItem[]>([]);
  const [scores, setScores] = useState<ScorePoint[]>([]);
  const [me, setMe] = useState<MeData | null>(null);

  useEffect(() => {
    const boot = async () => {
      const meRes = await fetch(`${API_BASE_URL}/v1/me`, {
        credentials: "include",
      });
      if (meRes.status === 401) {
        router.push("/login");
        return;
      }

      if (meRes.ok) {
        const meData = (await meRes.json()) as MeData;
        setMe(meData);
        if (!meData.onboardingSeen) {
          router.push("/welcome");
          return;
        }
      }

      try {
        const [runsRes, resumesRes] = await Promise.all([
          fetch(`${API_BASE_URL}/v1/runs?limit=20`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/v1/resumes?limit=20`, { credentials: "include" }),
        ]);

        const runList = runsRes.ok ? ((await runsRes.json()) as RunItem[]) : [];
        const resumeList = resumesRes.ok ? ((await resumesRes.json()) as ResumeItem[]) : [];

        setRuns(runList);
        setResumes(resumeList);

        const latestRuns = runList.slice(0, 8);
        const scoreResults = await Promise.allSettled(
          latestRuns.map(async (run) => {
            const runId = run.ID || run.id;
            if (!runId) return null;
            const reportRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}/report`, {
              credentials: "include",
            });
            if (!reportRes.ok) return null;
            const report = await reportRes.json();
            const atsRaw = report.ATSReport ?? report.atsReport ?? report.ats_report;
            const parsedScore = parseScore(atsRaw);
            const scoreValue =
              typeof parsedScore === "number" ? Math.round(parsedScore * 100) : null;
            if (scoreValue === null) return null;
            return { runId, score: scoreValue };
          }),
        );

        const parsedScores = scoreResults
          .map((result) => (result.status === "fulfilled" ? result.value : null))
          .filter((item): item is ScorePoint => Boolean(item));

        setScores(parsedScores.reverse());
      } catch {
        // non-fatal
      }
    };

    boot();
  }, [router]);

  const stats = useMemo(() => {
    const avgScore =
      scores.length > 0
        ? Math.round(scores.reduce((acc, point) => acc + point.score, 0) / scores.length)
        : 0;

    return [
      { label: "Runs completed", value: String(runs.length) },
      { label: "Base resumes", value: String(resumes.length) },
      { label: "Average ATS score", value: avgScore > 0 ? `${avgScore}%` : "-" },
    ];
  }, [runs.length, resumes.length, scores]);

  const chartPath = useMemo(() => {
    if (scores.length === 0) return "";
    const maxX = scores.length - 1 || 1;
    const width = 260;
    const height = 90;

    return scores
      .map((point, index) => {
        const x = (index / maxX) * width;
        const y = height - (point.score / 100) * height;
        return `${x},${y}`;
      })
      .join(" ");
  }, [scores]);

  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="private" />

        <main className="mt-4 space-y-4 sm:mt-6 sm:space-y-6">
          <section
            className="rt-panel rt-fade-up flex flex-col gap-4 p-5 sm:flex-row sm:items-end sm:justify-between sm:p-7"
            style={{ animationDelay: "40ms" }}
          >
            <div>
              <p className="rt-label">Dashboard</p>
              <h1 className="mt-2 font-grotesk text-4xl font-semibold text-[var(--rt-ink-900)] sm:text-5xl">
                Tailoring activity and momentum
              </h1>
              <p className="font-serif-display mt-2 text-xl text-[var(--rt-ink-700)]">
                Track what changed, which runs performed best, and where to tighten wording next.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              {me && !me.hasApiKey ? (
                <span className="rounded-full border border-[rgba(50,97,187,0.2)] bg-[rgba(216,228,247,0.72)] px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-[rgba(34,74,148,0.92)]">
                  {me.dailyRunsUsed} / {me.dailyRunsLimit} free runs today
                </span>
              ) : null}

              <button
                onClick={() => router.push("/resume")}
                className="rt-btn-primary px-6 py-3 text-xs font-semibold uppercase tracking-[0.22em]"
                aria-label="Start a new resume tailoring job"
              >
                Start a new run
              </button>
            </div>
          </section>

          <section className="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
            {/* Left: Activity panel */}
            <section className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "120ms" }}>
              <div className="flex items-start justify-between">
                <div>
                  <p className="rt-label">Activity</p>
                  <h3 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
                    Track every tailoring run.
                  </h3>
                </div>
                <button
                  onClick={() => router.push("/resume")}
                  className="rt-btn-primary shrink-0 px-4 py-2 text-xs font-semibold uppercase tracking-[0.22em]"
                >
                  Start a new run
                </button>
              </div>

              <div className="mt-4 space-y-2">
                {runs.length === 0 ? (
                  <div className="rounded-[1.1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] px-4 py-6 text-center">
                    <p className="text-sm text-[var(--rt-ink-700)]">No runs yet. Start your first tailoring run.</p>
                  </div>
                ) : (
                  runs.slice(0, 10).map((raw) => {
                    const run = normalizeRun(raw);
                    return (
                      <a
                        key={run.id}
                        href={`/result/${run.id}`}
                        className="flex items-center justify-between rounded-[1.1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] px-4 py-3 transition-colors hover:bg-[rgba(255,255,255,0.88)]"
                      >
                        <div className="min-w-0 flex-1">
                          <p className="truncate font-grotesk text-sm font-semibold text-[var(--rt-ink-900)]">
                            {run.jobTitle || "Untitled run"}{run.company ? ` · ${run.company}` : ""}
                          </p>

                          {run.createdAt && (
                            <p className="mt-0.5 text-xs text-[var(--rt-ink-500)]">
                              {formatRunDate(run.createdAt)}
                            </p>
                          )}
                        </div>
                        {run.status && (
                          <span className={`ml-3 shrink-0 rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase tracking-[0.1em] ${statusBadge(run.status)}`}>
                            {run.status.replace(/_/g, " ")}
                          </span>
                        )}
                      </a>
                    );
                  })
                )}
              </div>
            </section>

            {/* Right: Snapshot + ATS trend stacked */}
            <div className="flex flex-col gap-4">
              <section className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "160ms" }}>
                <p className="rt-label">Snapshot</p>
                <h3 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
                  At a glance
                </h3>
                <div className="mt-4 space-y-3">
                  {stats.map((stat) => (
                    <article
                      key={stat.label}
                      className="rounded-[1.1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.62)] px-4 py-4"
                    >
                      <p className="text-xs uppercase tracking-[0.18em] text-[var(--rt-ink-500)]">
                        {stat.label}
                      </p>
                      <p className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
                        {stat.value}
                      </p>
                    </article>
                  ))}
                </div>
              </section>

              <section className="rt-panel-muted rt-fade-up p-5 sm:p-6" style={{ animationDelay: "200ms" }}>
                <p className="rt-label">ATS trend</p>
                <h3 className="mt-2 font-grotesk text-3xl font-semibold text-[var(--rt-ink-900)]">
                  Score direction
                </h3>
                <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
                  Your latest BM25-backed scores on a 0-100 scale.
                </p>

                <div className="mt-4 rounded-[1.1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.58)] p-4">
                  {scores.length === 0 ? (
                    <p className="text-sm text-[var(--rt-ink-700)]">
                      Run a few jobs to reveal a trend line.
                    </p>
                  ) : (
                    <>
                      <svg
                        viewBox="0 0 260 90"
                        className="h-24 w-full"
                        role="img"
                        aria-label="ATS score trend chart"
                      >
                        <polyline
                          fill="none"
                          stroke="url(#trendGradient)"
                          strokeWidth="3"
                          strokeLinecap="round"
                          points={chartPath}
                        />
                        <defs>
                          <linearGradient id="trendGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" stopColor="#2e66d2" />
                            <stop offset="100%" stopColor="#2a8d6c" />
                          </linearGradient>
                        </defs>
                        {scores.map((point, index) => {
                          const maxX = scores.length - 1 || 1;
                          const x = (index / maxX) * 260;
                          const y = 90 - (point.score / 100) * 90;
                          return (
                            <circle
                              key={point.runId}
                              cx={x}
                              cy={y}
                              r="3"
                              fill="#2e66d2"
                              stroke="#fff"
                              strokeWidth="1.5"
                            />
                          );
                        })}
                      </svg>
                      <div className="mt-2 flex items-center justify-between text-xs uppercase tracking-[0.16em] text-[var(--rt-ink-500)]">
                        <span>Earlier</span>
                        <span>Latest</span>
                      </div>
                    </>
                  )}
                </div>
              </section>
            </div>
          </section>
        </main>
      </div>
    </div>
  );
}
