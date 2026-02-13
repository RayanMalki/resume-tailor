"use client";

import { useSearchParams, useRouter } from "next/navigation";
import { Suspense, useEffect, useMemo, useState } from "react";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

type ProjectMode = "pinned" | "auto" | "exclude";

type ProjectControl = {
  name: string;
  mode: ProjectMode;
};

function normalizeProjectName(input: string): string {
  return input
    .replace(/^[\-•*\d.()\s]+/, "")
    .replace(/\s+/g, " ")
    .trim();
}

function extractProjectCandidates(resumeText: string): string[] {
  const lines = resumeText
    .split(/\r?\n/)
    .map((line) => normalizeProjectName(line))
    .filter(Boolean);

  const candidates: string[] = [];
  for (const line of lines) {
    const lowered = line.toLowerCase();
    if (line.length < 4 || line.length > 90) continue;
    if (lowered.includes("experience") || lowered.includes("education") || lowered.includes("skills")) continue;

    if (line.includes("|") || line.includes(" — ") || line.includes(" - ")) {
      const name = normalizeProjectName(line.split(/[|—-]/)[0] || "");
      if (name.length >= 3) candidates.push(name);
      continue;
    }

    if (/[Pp]roject/.test(line) && line.includes(":")) {
      const name = normalizeProjectName(line.split(":")[1] || "");
      if (name.length >= 3) candidates.push(name);
      continue;
    }
  }

  const unique: string[] = [];
  const seen = new Set<string>();
  for (const c of candidates) {
    const key = c.toLowerCase();
    if (!seen.has(key)) {
      seen.add(key);
      unique.push(c);
    }
    if (unique.length >= 12) break;
  }
  return unique;
}

function JobPageInner() {
  const params = useSearchParams();
  const router = useRouter();
  const resumeId = params.get("resumeId");
  const [jobText, setJobText] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [projectControls, setProjectControls] = useState<ProjectControl[]>([]);
  const [newProjectName, setNewProjectName] = useState("");

  useEffect(() => {
    if (!resumeId) {
      router.push("/resume");
    }
  }, [resumeId, router]);

  useEffect(() => {
    const loadResumeProjects = async () => {
      if (!resumeId) return;
      try {
        const res = await fetch(`${API_BASE_URL}/v1/resumes/${resumeId}`, {
          credentials: "include"
        });
        if (!res.ok) return;
        const data = (await res.json()) as Record<string, unknown>;
        const contentText =
          (typeof data.ContentText === "string" && data.ContentText) ||
          (typeof data.contentText === "string" && data.contentText) ||
          "";
        if (!contentText) return;

        const detected = extractProjectCandidates(contentText);
        setProjectControls((prev) => {
          if (prev.length > 0) return prev;
          return detected.map((name) => ({ name, mode: "auto" as const }));
        });
      } catch {
        // Non-blocking: controls are optional.
      }
    };

    loadResumeProjects();
  }, [resumeId]);

  const sortedControls = useMemo(() => {
    return [...projectControls].sort((a, b) => a.name.localeCompare(b.name));
  }, [projectControls]);

  const setMode = (name: string, mode: ProjectMode) => {
    setProjectControls((prev) =>
      prev.map((p) => (p.name === name ? { ...p, mode } : p))
    );
  };

  const removeProject = (name: string) => {
    setProjectControls((prev) => prev.filter((p) => p.name !== name));
  };

  const addProject = () => {
    const name = normalizeProjectName(newProjectName);
    if (!name) return;
    setProjectControls((prev) => {
      if (prev.some((p) => p.name.toLowerCase() === name.toLowerCase())) {
        return prev;
      }
      return [...prev, { name, mode: "auto" }];
    });
    setNewProjectName("");
  };

  const handleGenerate = async () => {
    if (!resumeId) return;
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE_URL}/v1/runs`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify({
          resumeId,
          jobText,
          projectControls
        })
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

          <div className="mt-6 rounded-xl border border-white/10 bg-ink-950/40 p-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
              <div>
                <h2 className="text-sm font-semibold text-slate-200">Project relevance controls</h2>
                <p className="text-xs text-slate-400">Pin projects to force inclusion, exclude irrelevant ones, or keep auto.</p>
              </div>
            </div>

            <div className="mt-3 flex gap-2">
              <label htmlFor="add-project" className="sr-only">Add project name</label>
              <input
                id="add-project"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    addProject();
                  }
                }}
                className="w-full rounded-xl border border-white/10 bg-ink-950 px-3 py-2.5 text-sm text-slate-100 focus:border-ember-500/60 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                placeholder="Add project name"
              />
              <button
                onClick={addProject}
                className="rounded-xl border border-white/10 px-4 py-2.5 text-sm text-slate-100 transition hover:bg-white/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
                type="button"
              >
                Add
              </button>
            </div>

            <div className="mt-3 space-y-2">
              {sortedControls.length === 0 ? (
                <p className="text-xs text-slate-400">No projects detected yet. Add them manually if needed.</p>
              ) : (
                sortedControls.map((control) => (
                  <div
                    key={control.name}
                    className="flex flex-col gap-2 rounded-lg border border-white/10 bg-ink-900/60 px-3 py-2 sm:flex-row sm:items-center"
                  >
                    <span className="flex-1 text-sm text-slate-100">{control.name}</span>
                    <div className="flex items-center gap-2">
                      <label htmlFor={`mode-${control.name}`} className="sr-only">Mode for {control.name}</label>
                      <select
                        id={`mode-${control.name}`}
                        value={control.mode}
                        onChange={(e) => setMode(control.name, e.target.value as ProjectMode)}
                        className="rounded-md border border-white/10 bg-ink-950 px-2 py-1.5 text-xs text-slate-100 focus:outline-none focus:ring-1 focus:ring-ember-500/40"
                      >
                        <option value="pinned">Pin</option>
                        <option value="auto">Auto</option>
                        <option value="exclude">Exclude</option>
                      </select>
                      <button
                        onClick={() => removeProject(control.name)}
                        className="rounded-md border border-white/10 px-2 py-1.5 text-xs text-slate-300 transition hover:bg-white/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-ember-500"
                        type="button"
                        aria-label={`Remove ${control.name}`}
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
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
