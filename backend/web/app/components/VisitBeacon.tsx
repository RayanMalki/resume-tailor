"use client";

import { useEffect } from "react";
import { usePathname, useSearchParams } from "next/navigation";

export default function VisitBeacon() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const query = searchParams?.toString() ?? "";

  useEffect(() => {
    if (!pathname) return;

    const page = query ? `${pathname}?${query}` : pathname;
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
  }, [pathname, query]);

  return null;
}
