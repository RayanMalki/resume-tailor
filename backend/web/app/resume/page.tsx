"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";
import { useToast } from "../components/Toast";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

const ACCEPTED_FILE_TYPES = ".pdf,.docx";
const MAX_FILE_SIZE = 2 * 1024 * 1024; // 2 MB

export default function ResumePage() {
  const router = useRouter();
  const { toast } = useToast();
  const [error, setError] = useState<string | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadedFileName, setUploadedFileName] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const checkAuth = async () => {
      const res = await fetch(`${API_BASE_URL}/v1/me`, { credentials: "include" });
      if (res.status === 401) {
        router.push("/login");
      }
    };
    checkAuth();
  }, [router]);

  const handleFileUpload = useCallback(async (file: File) => {
    setError(null);

    // Validate file type
    const ext = file.name.toLowerCase().split(".").pop();
    if (ext !== "pdf" && ext !== "docx") {
      setError("Only PDF and DOCX files are accepted.");
      return;
    }

    // Validate file size
    if (file.size > MAX_FILE_SIZE) {
      setError("File is too large. Maximum size is 2 MB.");
      return;
    }

    setUploading(true);
    setUploadedFileName(file.name);

    try {
      const formData = new FormData();
      formData.append("file", file);

      const res = await fetch(`${API_BASE_URL}/v1/resumes/upload`, {
        method: "POST",
        headers: { "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: formData,
      });

      if (res.status === 401) {
        router.push("/login");
        return;
      }

      if (!res.ok) {
        const data = await res.json().catch(() => null);
        throw new Error(data?.error || "Failed to upload file");
      }

      const data = await res.json();
      toast("Resume uploaded successfully");
      router.push(`/job?resumeId=${data.resumeId}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to upload file";
      setError(msg);
      toast(msg, "error");
      setUploadedFileName(null);
    } finally {
      setUploading(false);
    }
  }, [router, toast]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragActive(false);
    const file = e.dataTransfer.files?.[0];
    if (file) handleFileUpload(file);
  }, [handleFileUpload]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragActive(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragActive(false);
  }, []);

  return (
    <div className="min-h-screen bg-ink-950 text-slate-100">
      <TopBar showLogout />
      <main className="mx-auto flex w-full max-w-5xl flex-1 items-start justify-center px-4 py-8 sm:px-6 sm:py-16">
        <div className="w-full max-w-2xl rounded-[28px] border border-white/10 bg-ink-900/70 p-5 shadow-panel backdrop-blur sm:p-8">
          <h1 className="text-xl font-semibold text-white sm:text-2xl">Upload your resume</h1>
          <p className="mt-2 text-sm text-slate-400">
            Upload a PDF or DOCX file. We&apos;ll extract the text and tailor it to your target job next.
          </p>

          {/* File upload drop zone */}
          <div
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onClick={() => fileInputRef.current?.click()}
            onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") fileInputRef.current?.click(); }}
            role="button"
            tabIndex={0}
            aria-label="Upload resume file — click or drag and drop"
            className={`mt-6 flex cursor-pointer flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed px-4 py-8 transition sm:py-10 ${
              dragActive
                ? "border-ember-500 bg-ember-500/10"
                : "border-white/10 bg-ink-950/40 hover:border-ember-500/40 hover:bg-ink-950/60"
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept={ACCEPTED_FILE_TYPES}
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) handleFileUpload(file);
                e.target.value = "";
              }}
              aria-hidden="true"
              tabIndex={-1}
            />
            {uploading ? (
              <>
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-ember-500/30 border-t-ember-500" />
                <p className="text-sm text-slate-300">Extracting text from {uploadedFileName}&hellip;</p>
              </>
            ) : (
              <>
                <svg className="h-8 w-8 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.338-2.32 3.75 3.75 0 013.57 4.595H18a3 3 0 01-.879 5.75M12 9.75v0" />
                </svg>
                <p className="text-sm text-slate-300">
                  {uploadedFileName ? (
                    <><span className="font-medium text-ember-300">{uploadedFileName}</span> uploaded</>
                  ) : (
                    <>Drag &amp; drop a <span className="font-medium text-ember-300">PDF</span> or <span className="font-medium text-ember-300">DOCX</span>, or click to browse</>
                  )}
                </p>
                <p className="text-xs text-slate-500">Max 2 MB</p>
              </>
            )}
          </div>

          {error ? (
            <div role="alert" className="mt-4 rounded-lg border border-rose-500/20 bg-rose-500/10 px-3 py-2">
              <p className="text-sm text-rose-300">{error}</p>
            </div>
          ) : null}
        </div>
      </main>
    </div>
  );
}
