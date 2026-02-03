import { NextRequest } from "next/server";

type VisitPayload = {
  page?: string;
  referrer?: string | null;
  ts?: string;
};

async function readPayload(request: NextRequest): Promise<VisitPayload> {
  const text = await request.text();
  if (!text) return {};

  try {
    return JSON.parse(text) as VisitPayload;
  } catch {
    return {};
  }
}

export async function POST(request: NextRequest) {
  const webhookUrl = process.env.DISCORD_WEBHOOK_URL;
  if (!webhookUrl) {
    return new Response(null, { status: 204 });
  }

  const payload = await readPayload(request);
  const page = payload.page || "unknown";
  const referrer = payload.referrer || "direct";
  const ts = payload.ts || new Date().toISOString();

  const content = [
    "New site visit",
    `Page: ${page}`,
    `Referrer: ${referrer}`,
    `Time: ${ts}`
  ].join("\n");

  await fetch(webhookUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content })
  });

  return new Response(null, { status: 204 });
}
