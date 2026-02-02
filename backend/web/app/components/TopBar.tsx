"use client";

import { useRouter } from "next/navigation";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export default function TopBar({ showLogout }: { showLogout?: boolean }) {
  const router = useRouter();

  const handleLogout = async () => {
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      credentials: "include"
    });
    router.push("/login");
  };

  return (
    <header className="w-full border-b border-slate-200 bg-white">
      <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-4">
        <div className="text-lg font-semibold text-slate-900">Resume Tailor</div>
        {showLogout ? (
          <button
            onClick={handleLogout}
            className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm text-slate-700 hover:border-slate-300"
          >
            Logout
          </button>
        ) : null}
      </div>
    </header>
  );
}
