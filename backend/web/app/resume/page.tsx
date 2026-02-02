"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export default function ResumePage() {
  const router = useRouter();
  const [resumeText, setResumeText] = useState("");
  const [title, setTitle] = useState("My Resume");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const checkAuth = async () => {
      const res = await fetch(`${API_BASE_URL}/v1/me`, { credentials: "include" });
      if (res.status === 401) {
        router.push("/login");
      }
    };
    checkAuth();
  }, [router]);

  const handleNext = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE_URL}/v1/resumes`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ title, contentText: resumeText })
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "Failed to create resume");
      }

      const data = await res.json();
      router.push(`/job?resumeId=${data.resumeId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create resume");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <TopBar showLogout />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-6 py-16">
        <div className="w-full max-w-2xl rounded-2xl bg-white p-8 shadow-card">
          <h1 className="text-xl font-semibold text-slate-900">Paste your resume</h1>
          <p className="mt-2 text-sm text-slate-600">
            Paste the full resume text. We will tailor it to the job description next.
          </p>

          <div className="mt-6 space-y-4">
            <div>
              <label className="text-sm font-medium text-slate-700">Title</label>
              <input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2"
              />
            </div>
            <div>
              <label className="text-sm font-medium text-slate-700">Resume text</label>
              <textarea
                value={resumeText}
                onChange={(e) => setResumeText(e.target.value)}
                rows={12}
                className="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2"
                placeholder="Paste your resume here..."
              />
            </div>
          </div>

          {error ? <p className="mt-4 text-sm text-red-600">{error}</p> : null}

          <div className="mt-6 flex justify-end">
            <button
              onClick={handleNext}
              disabled={loading || resumeText.trim().length === 0}
              className="rounded-lg bg-slate-900 px-4 py-2 text-white hover:bg-slate-800 disabled:opacity-70"
            >
              {loading ? "Saving..." : "Next"}
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}
