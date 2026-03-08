"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import TopBar from "../components/TopBar";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "/api";

interface MeResponse {
  userId: string;
  email: string;
  displayName: string;
  avatarUrl: string | null;
  authProvider: string;
  hasApiKey: boolean;
}

export default function SettingsPage() {
  const router = useRouter();
  const [me, setMe] = useState<MeResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // Change Password
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [passwordMsg, setPasswordMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const [passwordLoading, setPasswordLoading] = useState(false);

  // API Key
  const [apiKey, setApiKey] = useState("");
  const [apiKeyMsg, setApiKeyMsg] = useState<{ ok: boolean; text: string } | null>(null);
  const [apiKeyLoading, setApiKeyLoading] = useState(false);

  // Delete account
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteMsg, setDeleteMsg] = useState<string | null>(null);

  useEffect(() => {
    fetch(`${API_BASE_URL}/v1/me`, {
      headers: { "X-Requested-With": "XMLHttpRequest" },
      credentials: "include",
    })
      .then((r) => {
        if (!r.ok) { router.push("/login"); return null; }
        return r.json();
      })
      .then((data) => { if (data) setMe(data); })
      .catch(() => router.push("/login"))
      .finally(() => setLoading(false));
  }, [router]);

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordMsg(null);
    setPasswordLoading(true);
    try {
      const r = await fetch(`${API_BASE_URL}/v1/me/password`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify({ currentPassword, newPassword }),
      });
      if (r.ok) {
        setPasswordMsg({ ok: true, text: "Password updated successfully." });
        setCurrentPassword("");
        setNewPassword("");
      } else {
        const data = await r.json().catch(() => ({}));
        setPasswordMsg({ ok: false, text: data.error || "Failed to update password." });
      }
    } catch {
      setPasswordMsg({ ok: false, text: "Network error." });
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleSaveAPIKey = async (e: React.FormEvent) => {
    e.preventDefault();
    setApiKeyMsg(null);
    setApiKeyLoading(true);
    try {
      const r = await fetch(`${API_BASE_URL}/v1/me/api-key`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
        body: JSON.stringify({ apiKey }),
      });
      if (r.ok) {
        setApiKeyMsg({ ok: true, text: "API key saved." });
        setApiKey("");
        setMe((prev) => prev ? { ...prev, hasApiKey: true } : prev);
      } else {
        const data = await r.json().catch(() => ({}));
        setApiKeyMsg({ ok: false, text: data.error || "Failed to save API key." });
      }
    } catch {
      setApiKeyMsg({ ok: false, text: "Network error." });
    } finally {
      setApiKeyLoading(false);
    }
  };

  const handleRemoveAPIKey = async () => {
    setApiKeyMsg(null);
    setApiKeyLoading(true);
    try {
      const r = await fetch(`${API_BASE_URL}/v1/me/api-key`, {
        method: "DELETE",
        headers: { "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
      });
      if (r.ok) {
        setApiKeyMsg({ ok: true, text: "API key removed." });
        setMe((prev) => prev ? { ...prev, hasApiKey: false } : prev);
      } else {
        setApiKeyMsg({ ok: false, text: "Failed to remove API key." });
      }
    } catch {
      setApiKeyMsg({ ok: false, text: "Network error." });
    } finally {
      setApiKeyLoading(false);
    }
  };

  const handleDeleteAccount = async () => {
    setDeleteLoading(true);
    setDeleteMsg(null);
    try {
      const r = await fetch(`${API_BASE_URL}/v1/me`, {
        method: "DELETE",
        headers: { "X-Requested-With": "XMLHttpRequest" },
        credentials: "include",
      });
      if (r.ok) {
        router.push("/login");
      } else {
        setDeleteMsg("Failed to delete account. Please try again.");
        setDeleteLoading(false);
      }
    } catch {
      setDeleteMsg("Network error.");
      setDeleteLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-ink-950 flex items-center justify-center">
        <div className="text-slate-400 text-sm">Loading…</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-ink-950 text-white">
      <TopBar showLogout />
      <main className="mx-auto max-w-2xl px-4 py-12 sm:px-6">
        <h1 className="mb-8 text-2xl font-bold text-white">Settings</h1>

        {/* Change Password — local auth only */}
        {me?.authProvider === "local" && (
          <section className="mb-8 rounded-2xl border border-white/10 bg-ink-900/60 p-6">
            <h2 className="mb-4 text-base font-semibold text-white">Change Password</h2>
            <form onSubmit={handleChangePassword} className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-xs text-slate-400">Current password</label>
                <input
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  className="w-full rounded-xl border border-white/10 bg-ink-950 px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:border-ember-500 focus:outline-none"
                  required
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-slate-400">New password</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full rounded-xl border border-white/10 bg-ink-950 px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:border-ember-500 focus:outline-none"
                  required
                />
              </div>
              {passwordMsg && (
                <p className={`text-xs ${passwordMsg.ok ? "text-green-400" : "text-red-400"}`}>
                  {passwordMsg.text}
                </p>
              )}
              <button
                type="submit"
                disabled={passwordLoading}
                className="self-start rounded-full border border-ember-500/60 bg-ink-900 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-ember-200 transition hover:border-ember-400 hover:text-white disabled:opacity-50"
              >
                {passwordLoading ? "Saving…" : "Update Password"}
              </button>
            </form>
          </section>
        )}

        {/* API Key */}
        <section className="mb-8 rounded-2xl border border-white/10 bg-ink-900/60 p-6">
          <h2 className="mb-1 text-base font-semibold text-white">OpenAI API Key</h2>
          <p className="mb-4 text-xs text-slate-400">
            Your key is encrypted server-side and used instead of the shared key when you run an analysis. The model remains unchanged.
          </p>

          {/* Security disclaimer */}
          <div className="mb-5 rounded-xl border border-white/10 bg-ink-950/60 p-4">
            <p className="mb-2 text-xs font-semibold text-slate-300">
              🔒 How we protect your key
            </p>
            <p className="text-xs text-slate-400 leading-relaxed">
              Your key is encrypted with <span className="text-slate-300">AES-256-GCM</span> before
              it is stored. The encryption secret lives only on the server — never in the database.
              Even a complete database dump cannot expose your key. It is decrypted in memory only
              during an analysis run and is never sent back to your browser.
            </p>
          </div>

          {me?.hasApiKey && (
            <p className="mb-3 text-xs text-green-400">A personal API key is currently saved.</p>
          )}
          <form onSubmit={handleSaveAPIKey} className="flex flex-col gap-4">
            <div>
              <label className="mb-1 block text-xs text-slate-400">
                {me?.hasApiKey ? "Replace with new key" : "Paste your OpenAI API key"}
              </label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-..."
                className="w-full rounded-xl border border-white/10 bg-ink-950 px-4 py-2.5 text-sm text-white placeholder-slate-500 focus:border-ember-500 focus:outline-none"
                required
              />
            </div>
            {apiKeyMsg && (
              <p className={`text-xs ${apiKeyMsg.ok ? "text-green-400" : "text-red-400"}`}>
                {apiKeyMsg.text}
              </p>
            )}
            <div className="flex gap-3">
              <button
                type="submit"
                disabled={apiKeyLoading}
                className="rounded-full border border-ember-500/60 bg-ink-900 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-ember-200 transition hover:border-ember-400 hover:text-white disabled:opacity-50"
              >
                {apiKeyLoading ? "Saving…" : "Save Key"}
              </button>
              {me?.hasApiKey && (
                <button
                  type="button"
                  onClick={handleRemoveAPIKey}
                  disabled={apiKeyLoading}
                  className="rounded-full border border-white/10 bg-ink-900 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-slate-300 transition hover:border-red-400/60 hover:text-red-300 disabled:opacity-50"
                >
                  Remove Key
                </button>
              )}
            </div>
          </form>
        </section>

        {/* Danger Zone */}
        <section className="rounded-2xl border border-red-500/20 bg-ink-900/60 p-6">
          <h2 className="mb-1 text-base font-semibold text-red-400">Danger Zone</h2>
          <p className="mb-4 text-xs text-slate-400">
            Permanently delete your account and all associated data. This action cannot be undone.
          </p>
          {deleteMsg && <p className="mb-3 text-xs text-red-400">{deleteMsg}</p>}
          {!showDeleteConfirm ? (
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="rounded-full border border-red-500/40 bg-ink-950 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-red-400 transition hover:border-red-400 hover:text-red-300"
            >
              Delete My Account
            </button>
          ) : (
            <div className="flex flex-col gap-3">
              <p className="text-sm font-semibold text-red-300">Are you sure? This cannot be undone.</p>
              <div className="flex gap-3">
                <button
                  onClick={handleDeleteAccount}
                  disabled={deleteLoading}
                  className="rounded-full border border-red-500/60 bg-red-900/30 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-red-300 transition hover:border-red-400 hover:text-red-200 disabled:opacity-50"
                >
                  {deleteLoading ? "Deleting…" : "Yes, Delete Everything"}
                </button>
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  disabled={deleteLoading}
                  className="rounded-full border border-white/10 bg-ink-900 px-5 py-2 text-xs font-semibold uppercase tracking-[0.15em] text-slate-300 transition hover:border-white/30 disabled:opacity-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
