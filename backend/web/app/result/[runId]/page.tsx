"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import TopBar from "../../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export default function ResultPage() {
  const router = useRouter();
  const params = useParams<{ runId: string }>();
  const runId = params.runId;
  const [latex, setLatex] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState(10);
  const [resumeId, setResumeId] = useState<string | null>(null);

  const loadingMessage = useMemo(() => {
    if (latex) return "Ready";
    return "Generating LaTeX...";
  }, [latex]);

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

  useEffect(() => {
    let poller: NodeJS.Timeout;
    const extractResumeId = (data: Record<string, unknown>) => {
      const raw =
        (data.resumeId as string | undefined) ||
        (data.ResumeID as string | undefined) ||
        (data.resume_id as string | undefined);
      if (raw) {
        setResumeId((prev) => prev || raw);
      }
    };
    const fetchRunMeta = async () => {
      try {
        const runRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}`, {
          credentials: "include"
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
        // ignore metadata failures; resumeId is optional for navigation
      }
    };
    const poll = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/runs/${runId}/artifacts/resume-latex`, {
          credentials: "include"
        });

        if (res.status === 401) {
          router.push("/login");
          return;
        }

        if (res.ok) {
          const data = await res.json();
          setLatex(data.latex);
          setProgress(100);
          return;
        }

        if (res.status === 404) {
          const runRes = await fetch(`${API_BASE_URL}/v1/runs/${runId}`, {
            credentials: "include"
          });
          if (runRes.ok) {
            const runData = await runRes.json();
            extractResumeId(runData);
            if (runData.status === "failed") {
              setError(runData.errorMessage || "Run failed");
            }
          }
          return;
        }

        setError("Failed to fetch artifact");
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch artifact");
      }
    };

    poller = setInterval(poll, 1000);
    poll();
    fetchRunMeta();

    return () => clearInterval(poller);
  }, [runId, router]);

  const handleCopy = async () => {
    if (!latex) return;
    await navigator.clipboard.writeText(latex);
  };

  const handleGoToResume = () => {
    router.push("/resume");
  };

  const handleGoToJob = () => {
    if (resumeId) {
      router.push(`/job?resumeId=${resumeId}`);
      return;
    }
    router.push("/job");
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <TopBar showLogout />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-6 py-16">
        <div className="w-full max-w-3xl rounded-2xl bg-white p-8 shadow-card">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h1 className="text-xl font-semibold text-slate-900">Your tailored LaTeX</h1>
            <div className="flex flex-wrap items-center gap-2">
              <button
                onClick={handleGoToResume}
                className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-300"
              >
                Upload new CV
              </button>
              <button
                onClick={handleGoToJob}
                className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-300"
              >
                Upload new job listing
              </button>
            </div>
          </div>
          <p className="mt-2 text-sm text-slate-600">{loadingMessage}</p>

          {!latex ? (
            <div className="mt-6">
              <div className="h-2 w-full overflow-hidden rounded-full bg-slate-100">
                <div
                  className="h-full rounded-full bg-slate-900 transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
              {error ? <p className="mt-4 text-sm text-red-600">{error}</p> : null}
            </div>
          ) : (
            <div className="mt-6">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-slate-700">LaTeX output</span>
                <button
                  onClick={handleCopy}
                  className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-300"
                >
                  Copy
                </button>
              </div>
              <pre className="mt-3 max-h-[420px] overflow-auto rounded-lg bg-slate-900 p-4 text-xs text-slate-100">
{latex}
              </pre>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
