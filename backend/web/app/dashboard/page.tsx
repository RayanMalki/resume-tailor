"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import EditorialNav from "../components/EditorialNav";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

type RunItem = {
  ID?: string;
  id?: string;
  ResumeID?: string;
  resumeId?: string;
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

          <section className="grid gap-4 lg:grid-cols-2">
            <section className="rt-panel rt-fade-up p-5 sm:p-6" style={{ animationDelay: "120ms" }}>
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

            <section className="rt-panel-muted rt-fade-up p-5 sm:p-6" style={{ animationDelay: "180ms" }}>
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
          </section>
        </main>
      </div>
    </div>
  );
}
