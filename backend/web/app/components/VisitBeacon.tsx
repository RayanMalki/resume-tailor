"use client";

import { useEffect } from "react";

export default function VisitBeacon() {
  useEffect(() => {
    if (typeof window === "undefined") return;

    const { pathname, search } = window.location;
    const page = search ? `${pathname}${search}` : pathname;
    const payload = JSON.stringify({
      page,
      referrer: document.referrer || null,
      ts: new Date().toISOString()
    });

    if (navigator.sendBeacon) {
      navigator.sendBeacon("/api/visit", payload);
      return;
    }

    fetch("/api/visit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: payload,
      keepalive: true
    }).catch(() => {});
  }, []);

  return null;
}
