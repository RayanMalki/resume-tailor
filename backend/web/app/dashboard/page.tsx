"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type RunItem = {
  ID?: string;
  id?: string;
  ResumeID?: string;
  resumeId?: string;
  JobText?: string;
  jobText?: string;
  Status?: string;
  status?: string;
  CreatedAt?: string;
  createdAt?: string;
};

type ResumeItem = {
  ID?: string;
  id?: string;
  Title?: string;
  title?: string;
};

type ScorePoint = {
  runId: string;
  score: number;
};

const parseScore = (atsRaw: unknown): number | null => {
  const parsed = typeof atsRaw === "string" ? (() => {
    try {
      return JSON.parse(atsRaw);
    } catch {
      return null;
    }
  })() : atsRaw;

  if (!parsed || typeof parsed !== "object") return null;

  const asRecord = parsed as Record<string, unknown>;
  const directScore = asRecord.score;
  if (typeof directScore === "number") return directScore;

  const nested =
    asRecord.ats_report ||
    asRecord.atsReport ||
    asRecord.ATSReport;
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
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const boot = async () => {
      const meRes = await fetch(`${API_BASE_URL}/v1/me`, { credentials: "include" });
      if (meRes.status === 401) {
        router.push("/login");
        return;
      }

      try {
        const [runsRes, resumesRes] = await Promise.all([
          fetch(`${API_BASE_URL}/v1/runs?limit=20`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/v1/resumes?limit=20`, { credentials: "include" })
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
              credentials: "include"
            });
            if (!reportRes.ok) return null;
            const report = await reportRes.json();
            const atsRaw = report.ATSReport ?? report.atsReport ?? report.ats_report;
            const parsedScore = parseScore(atsRaw);
            const scoreValue =
              typeof parsedScore === "number" ? Math.round(parsedScore * 100) : null;
            if (scoreValue === null) return null;
            return { runId, score: scoreValue };
          })
        );

        const parsedScores = scoreResults
          .map((result) => (result.status === "fulfilled" ? result.value : null))
          .filter((item): item is ScorePoint => Boolean(item));

        setScores(parsedScores.reverse());
      } finally {
        setLoading(false);
      }
    };

    boot();
  }, [router]);

  const stats = useMemo(() => {
    const jobCount = runs.length;
    const resumeCount = resumes.length;
    return [
      { label: "Jobs analyzed", value: jobCount.toString() },
      { label: "CVs uploaded", value: resumeCount.toString() },
      { label: "Job listings pasted", value: jobCount.toString() }
    ];
  }, [runs.length, resumes.length]);

  const chartPath = useMemo(() => {
    if (scores.length === 0) return "";
    const maxX = scores.length - 1 || 1;
    const width = 240;
    const height = 80;
    return scores
      .map((point, index) => {
        const x = (index / maxX) * width;
        const y = height - (point.score / 100) * height;
        return `${x},${y}`;
      })
      .join(" ");
  }, [scores]);

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar showLogout />
      <main className="mx-auto w-full max-w-6xl px-6 py-10">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.35em] text-ember-300/80">
              Dashboard
            </p>
            <h1 className="mt-2 text-3xl font-semibold text-white">
              Track your tailored runs.
            </h1>
          </div>
          <button
            onClick={() => router.push("/resume")}
            className="rounded-full bg-ember-500 px-6 py-2.5 text-sm font-semibold text-ink-950 shadow-glow transition hover:-translate-y-0.5 hover:bg-ember-400"
          >
            Start a new job
          </button>
        </div>

        <section className="mt-8 grid gap-5 lg:grid-cols-[2fr_1fr]">
          <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-ember-300/80">
                  Activity
                </p>
                <h2 className="mt-2 text-lg font-semibold text-white">Your job runs</h2>
              </div>
              <span className="rounded-full border border-white/10 px-3 py-1 text-xs text-slate-300">
                {runs.length} total
              </span>
            </div>

            <div className="mt-6 space-y-3">
              {loading ? (
                <div className="rounded-2xl border border-white/10 bg-ink-950/70 px-4 py-6 text-sm text-slate-400">
                  Loading runs...
                </div>
              ) : runs.length === 0 ? (
                <div className="rounded-2xl border border-white/10 bg-ink-950/70 px-4 py-6 text-sm text-slate-400">
                  No runs yet. Start your first job to see it here.
                </div>
              ) : (
                runs.map((run) => {
                  const runId = run.ID || run.id || "";
                  const status = (run.Status || run.status || "queued").toString();
                  const createdRaw = run.CreatedAt || run.createdAt;
                  const created = createdRaw
                    ? new Date(createdRaw).toLocaleString()
                    : "Just now";
                  const jobText = (run.JobText || run.jobText || "").replace(/\s+/g, " ").trim();
                  const snippet = jobText ? `${jobText.slice(0, 80)}${jobText.length > 80 ? "…" : ""}` : "Job listing";

                  return (
                    <button
                      key={runId}
                      onClick={() => router.push(`/result/${runId}`)}
                      className="flex w-full items-center justify-between rounded-2xl border border-white/10 bg-ink-950/70 px-4 py-4 text-left transition hover:border-ember-400/60 hover:bg-ink-950"
                    >
                      <div>
                        <p className="text-sm font-semibold text-white">{snippet}</p>
                        <p className="mt-1 text-xs text-slate-400">{created}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="rounded-full border border-ember-500/60 px-3 py-1 text-[11px] uppercase tracking-[0.2em] text-ember-300">
                          {status}
                        </span>
                        <span className="text-xs text-ember-300">View</span>
                      </div>
                    </button>
                  );
                })
              )}
            </div>
          </div>

          <div className="space-y-5">
            <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur">
              <p className="text-xs uppercase tracking-[0.3em] text-ember-300/80">Stats</p>
              <div className="mt-5 space-y-4">
                {stats.map((stat) => (
                  <div
                    key={stat.label}
                    className="rounded-2xl border border-white/10 bg-ink-950/80 px-4 py-4"
                  >
                    <p className="text-xs uppercase tracking-[0.2em] text-slate-400">
                      {stat.label}
                    </p>
                    <p className="mt-2 text-2xl font-semibold text-white">{stat.value}</p>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-white/10 bg-ink-900/70 p-6 shadow-panel backdrop-blur">
              <p className="text-xs uppercase tracking-[0.3em] text-ember-300/80">
                ATS score trend
              </p>
              <p className="mt-2 text-sm text-slate-400">
                Latest BM25-backed ATS scores (0-100).
              </p>
              <div className="mt-4 rounded-2xl border border-white/10 bg-ink-950/80 px-4 py-4">
                {scores.length === 0 ? (
                  <p className="text-xs text-slate-400">
                    Run a few jobs to see the trend line.
                  </p>
                ) : (
                  <svg viewBox="0 0 240 80" className="h-24 w-full">
                    <polyline
                      fill="none"
                      stroke="#ff7a1a"
                      strokeWidth="3"
                      points={chartPath}
                    />
                    {scores.map((point, index) => {
                      const maxX = scores.length - 1 || 1;
                      const x = (index / maxX) * 240;
                      const y = 80 - (point.score / 100) * 80;
                      return (
                        <circle key={point.runId} cx={x} cy={y} r="3" fill="#ffb57f" />
                      );
                    })}
                  </svg>
                )}
                <div className="mt-3 flex items-center justify-between text-xs text-slate-500">
                  <span>Earlier</span>
                  <span>Latest</span>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
