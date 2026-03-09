"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import EditorialNav from "../components/EditorialNav";

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
      <div className="rt-canvas flex items-center justify-center">
        <p className="text-sm text-[var(--rt-ink-500)]">Loading…</p>
      </div>
    );
  }

  return (
    <div className="rt-canvas">
      <div className="rt-shell">
        <EditorialNav mode="private" />
        <main className="mt-4 space-y-4 sm:mt-6 sm:space-y-6">

          {/* Header */}
          <section className="rt-panel rt-fade-up p-5 sm:p-7">
            <p className="rt-label">Account</p>
            <h1 className="mt-2 font-grotesk text-4xl font-semibold text-[var(--rt-ink-900)] sm:text-5xl">
              Settings
            </h1>
            {me?.email && (
              <p className="font-serif-display mt-2 text-xl text-[var(--rt-ink-700)]">
                {me.email}
              </p>
            )}
          </section>

          {/* Change Password — local auth only */}
          {me?.authProvider === "local" && (
            <section className="rt-panel rt-fade-up p-5 sm:p-7" style={{ animationDelay: "60ms" }}>
              <p className="rt-label">Security</p>
              <h2 className="mt-2 font-grotesk text-2xl font-semibold text-[var(--rt-ink-900)]">
                Change Password
              </h2>
              <form onSubmit={handleChangePassword} className="mt-5 flex flex-col gap-4">
                <div>
                  <label className="mb-1.5 block text-xs font-semibold uppercase tracking-[0.14em] text-[var(--rt-ink-500)]">
                    Current password
                  </label>
                  <input
                    type="password"
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    className="w-full rounded-[0.9rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.72)] px-4 py-3 text-sm text-[var(--rt-ink-900)] placeholder-[var(--rt-ink-400)] focus:border-[var(--rt-blue)] focus:outline-none"
                    required
                  />
                </div>
                <div>
                  <label className="mb-1.5 block text-xs font-semibold uppercase tracking-[0.14em] text-[var(--rt-ink-500)]">
                    New password
                  </label>
                  <input
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    className="w-full rounded-[0.9rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.72)] px-4 py-3 text-sm text-[var(--rt-ink-900)] placeholder-[var(--rt-ink-400)] focus:border-[var(--rt-blue)] focus:outline-none"
                    required
                  />
                </div>
                {passwordMsg && (
                  <p className={`text-sm ${passwordMsg.ok ? "text-[var(--rt-green)]" : "text-red-500"}`}>
                    {passwordMsg.text}
                  </p>
                )}
                <div>
                  <button
                    type="submit"
                    disabled={passwordLoading}
                    className="rt-btn-primary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] disabled:opacity-50"
                  >
                    {passwordLoading ? "Saving…" : "Update Password"}
                  </button>
                </div>
              </form>
            </section>
          )}

          {/* API Key */}
          <section className="rt-panel rt-fade-up p-5 sm:p-7" style={{ animationDelay: "120ms" }}>
            <p className="rt-label">Integrations</p>
            <h2 className="mt-2 font-grotesk text-2xl font-semibold text-[var(--rt-ink-900)]">
              OpenAI API Key
            </h2>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              Your key is encrypted server-side and used instead of the shared key when you run an analysis.
            </p>

            {/* Security note */}
            <div className="mt-5 rounded-[1rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.58)] p-4">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[var(--rt-ink-500)]">
                How we protect your key
              </p>
              <p className="mt-2 text-sm leading-relaxed text-[var(--rt-ink-700)]">
                Your key is encrypted with <span className="font-semibold text-[var(--rt-ink-900)]">AES-256-GCM</span> before
                it is stored. The encryption secret lives only on the server — never in the database.
                Even a complete database dump cannot expose your key. It is decrypted in memory only
                during an analysis run and is never sent back to your browser.
              </p>
            </div>

            {me?.hasApiKey && (
              <p className="mt-4 text-sm text-[var(--rt-green)]">A personal API key is currently saved.</p>
            )}

            <form onSubmit={handleSaveAPIKey} className="mt-5 flex flex-col gap-4">
              <div>
                <label className="mb-1.5 block text-xs font-semibold uppercase tracking-[0.14em] text-[var(--rt-ink-500)]">
                  {me?.hasApiKey ? "Replace with new key" : "Paste your OpenAI API key"}
                </label>
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="sk-..."
                  className="w-full rounded-[0.9rem] border border-[var(--rt-stroke)] bg-[rgba(255,255,255,0.72)] px-4 py-3 text-sm text-[var(--rt-ink-900)] placeholder-[var(--rt-ink-400)] focus:border-[var(--rt-blue)] focus:outline-none"
                  required
                />
              </div>
              {apiKeyMsg && (
                <p className={`text-sm ${apiKeyMsg.ok ? "text-[var(--rt-green)]" : "text-red-500"}`}>
                  {apiKeyMsg.text}
                </p>
              )}
              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={apiKeyLoading}
                  className="rt-btn-primary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] disabled:opacity-50"
                >
                  {apiKeyLoading ? "Saving…" : "Save Key"}
                </button>
                {me?.hasApiKey && (
                  <button
                    type="button"
                    onClick={handleRemoveAPIKey}
                    disabled={apiKeyLoading}
                    className="rt-btn-secondary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] disabled:opacity-50"
                  >
                    Remove Key
                  </button>
                )}
              </div>
            </form>
          </section>

          {/* Danger Zone */}
          <section className="rt-panel rt-fade-up p-5 sm:p-7" style={{ animationDelay: "180ms" }}>
            <p className="rt-label" style={{ color: "rgba(185,28,28,0.7)" }}>Danger Zone</p>
            <h2 className="mt-2 font-grotesk text-2xl font-semibold text-[var(--rt-ink-900)]">
              Delete Account
            </h2>
            <p className="font-serif-display mt-2 text-lg text-[var(--rt-ink-700)]">
              Permanently delete your account and all associated data. This action cannot be undone.
            </p>
            {deleteMsg && <p className="mt-3 text-sm text-red-500">{deleteMsg}</p>}
            <div className="mt-5">
              {!showDeleteConfirm ? (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="rt-btn-secondary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-red-500 hover:border-red-300"
                >
                  Delete My Account
                </button>
              ) : (
                <div className="flex flex-col gap-4">
                  <p className="text-sm font-semibold text-[var(--rt-ink-900)]">Are you sure? This cannot be undone.</p>
                  <div className="flex gap-3">
                    <button
                      onClick={handleDeleteAccount}
                      disabled={deleteLoading}
                      className="rt-btn-secondary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-red-500 hover:border-red-300 disabled:opacity-50"
                    >
                      {deleteLoading ? "Deleting…" : "Yes, Delete Everything"}
                    </button>
                    <button
                      onClick={() => setShowDeleteConfirm(false)}
                      disabled={deleteLoading}
                      className="rt-btn-secondary px-6 py-2.5 text-xs font-semibold uppercase tracking-[0.18em] disabled:opacity-50"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              )}
            </div>
          </section>

        </main>
      </div>
    </div>
  );
}
