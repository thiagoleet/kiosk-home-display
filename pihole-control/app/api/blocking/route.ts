import type { BlockingStatus } from "../../blocking-status";

const noStore = { "Cache-Control": "no-store" };
const pauseSeconds = 300;

function reply(data: BlockingStatus | { error: string }, status = 200) {
  return Response.json(data, { headers: noStore, status });
}

async function blocking(method: "GET" | "POST") {
  let origin: string;
  try {
    const url = new URL(process.env.PIHOLE_URL ?? "");
    if (!["http:", "https:"].includes(url.protocol) || url.username || url.password ||
        url.pathname !== "/" || url.search || url.hash) {
      throw new Error("Invalid Pi-hole origin");
    }
    origin = url.origin;
  } catch {
    return reply({ error: "Set PIHOLE_URL to the origin of your Pi-hole v6 server." }, 503);
  }

  let sid = "";
  function request(path: string, method: string, payload?: object) {
    const headers = new Headers();
    if (sid) headers.set("X-FTL-SID", sid);
    if (payload) headers.set("Content-Type", "application/json");
    return fetch(origin + path, {
      method,
      headers,
      body: payload ? JSON.stringify(payload) : undefined,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(5000),
    });
  }

  try {
    const password = process.env.PIHOLE_PASSWORD;
    if (password) {
      const response = await request("/api/auth", "POST", { password });
      if (!response.ok) {
        return reply({ error: "Pi-hole sign-in failed. Check the application password." }, 502);
      }
      const { session } = await response.json();
      if (!session?.valid || typeof session.sid !== "string" || !session.sid) {
        return reply({ error: "Pi-hole did not provide a valid API session." }, 502);
      }
      sid = session.sid;
    }

    const response = await request("/api/dns/blocking", method,
      method === "POST" ? { blocking: false, timer: pauseSeconds } : undefined);
    if (!response.ok) throw new Error("Pi-hole status request failed");
    const { blocking, timer } = await response.json();
    if (blocking !== "enabled" && blocking !== "disabled") {
      throw new Error("Unexpected blocking status");
    }
    const remaining = typeof timer === "number" && timer > 0 ? timer : null;
    if (method === "POST" && (blocking !== "disabled" || remaining === null)) {
      throw new Error("Timed pause was not confirmed");
    }
    return reply({ blocking, timer: remaining });
  } catch {
    return reply({ error: method === "POST"
      ? "Pi-hole could not confirm the pause. Check blocking status before retrying."
      : "Could not read Pi-hole blocking status. Check the connection and credentials."
    }, 502);
  } finally {
    if (sid) {
      try {
        const response = await request("/api/auth", "DELETE");
        if (!response.ok) console.warn("Could not release the Pi-hole API session");
      } catch {
        console.warn("Could not release the Pi-hole API session");
      }
    }
  }
}

export function GET() {
  return blocking("GET");
}

export function POST(request: Request) {
  const host = request.headers.get("host");
  const origin = request.headers.get("origin");
  const allowedHosts = ["milkpi.local", "127.0.0.1", "localhost", "127.0.0.1:8081", "localhost:8081"];
  if (!host || !allowedHosts.includes(host) ||
      origin !== `http://${host}` ||
      request.headers.get("X-Pihole-Control") !== "disable") {
    return reply({ error: "Request rejected." }, 403);
  }
  return blocking("POST");
}
