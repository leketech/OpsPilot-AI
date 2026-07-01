import { NextRequest, NextResponse } from "next/server";

// Read at runtime from the container env — set via docker-compose environment:
// BACKEND_URL=http://api:8080 (Docker) or falls back to localhost (local dev).
const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

// Vercel: lift the default 30s limit. No-op in local Docker but documents intent.
export const maxDuration = 300;

export async function POST(request: NextRequest) {
  const body = await request.text();

  // AbortController gives us a hard ceiling so the connection is always cleaned up.
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 290_000); // 290 s

  try {
    const upstream = await fetch(
      `${BACKEND_URL}/api/v1/incidents/analyze`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body,
        signal: controller.signal,
        cache: "no-store",
      }
    );

    const text = await upstream.text();
    return new NextResponse(text, {
      status: upstream.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch (err) {
    if (err instanceof Error && err.name === "AbortError") {
      return NextResponse.json(
        { error: "Analysis timed out — pipeline took more than 290 s" },
        { status: 504 }
      );
    }
    return NextResponse.json(
      {
        error: `Cannot reach backend: ${
          err instanceof Error ? err.message : String(err)
        }`,
      },
      { status: 502 }
    );
  } finally {
    clearTimeout(timer);
  }
}
