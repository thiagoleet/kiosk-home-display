"use client";

import { PixelButton } from "@pxlkit/ui-kit";
import { useCallback, useEffect, useRef, useState } from "react";
import type { BlockingStatus } from "./blocking-status";

export default function BlockingControl() {
  const [status, setStatus] = useState<BlockingStatus>();
  const [readError, setReadError] = useState("");
  const [actionError, setActionError] = useState("");
  const [pausing, setPausing] = useState(false);
  const currentRequest = useRef<AbortController | null>(null);

  const update = useCallback(async (method: "GET" | "POST") => {
    if (currentRequest.current && method === "GET") return;
    currentRequest.current?.abort();
    const controller = new AbortController();
    currentRequest.current = controller;
    if (method === "POST") {
      setPausing(true);
      setActionError("");
    }

    try {
      const response = await fetch("/api/blocking", {
        method,
        signal: controller.signal,
        headers: method === "POST" ? { "X-Pihole-Control": "disable" } : {},
      });
      const data = await response.json();
      if (controller.signal.aborted) return;
      if (!response.ok) throw new Error(data.error);
      setStatus(data);
      setReadError("");
      if (method === "GET" && data.blocking === "disabled" && data.timer) {
        setActionError("");
      }
    } catch (error) {
      if (controller.signal.aborted) return;
      const message = error instanceof Error && !(error instanceof TypeError)
        ? error.message
        : "Could not reach Pi-hole Control.";
      setStatus(undefined);
      if (method === "POST") setActionError(message);
      else setReadError(message);
    } finally {
      if (currentRequest.current === controller) {
        currentRequest.current = null;
        if (method === "POST") setPausing(false);
      }
    }
  }, []);

  useEffect(() => {
    void update("GET");
    const poll = window.setInterval(() => void update("GET"), 5000);
    return () => {
      window.clearInterval(poll);
      currentRequest.current?.abort();
      currentRequest.current = null;
    };
  }, [update]);

  let summary = readError ? "Blocking status unavailable" : "Checking Pi-hole…";
  if (status?.blocking === "enabled") summary = "Blocking enabled";
  if (status?.blocking === "disabled") {
    summary = status.timer
      ? `Blocking paused · about ${Math.ceil(status.timer / 60)} min left`
      : "Blocking disabled";
  }
  if (pausing) summary = "Pausing blocking…";
  else if (actionError && !status) summary = "Checking blocking status…";

  return (
    <section className="blocking-panel" aria-labelledby="blocking-heading">
      <h2 id="blocking-heading">DNS blocking</h2>
      <p role="status" className="blocking-status" data-state={status?.blocking}>{summary}</p>
      <PixelButton
        type="button"
        tone="green"
        className="pause-button"
        disabled={pausing || status?.blocking !== "enabled"}
        onClick={() => void update("POST")}
      >
        Disable Block for 5 minutes
      </PixelButton>
      <p className="hint">Pi-hole automatically resumes blocking after five minutes.</p>
      {(actionError || readError) && <p className="error" role="alert">{actionError || readError}</p>}
    </section>
  );
}
