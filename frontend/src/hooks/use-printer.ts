import { useCallback, useEffect, useState } from "react";

import { useWebSocketContext } from "./use-websocket-context";

import type { PrintJob } from "../types/printer";
import type { WebSocketMessage } from "../types/websocket";

export type PrinterState = "idle" | "printing" | "completed";

export function usePrinter() {
  const [state, setState] = useState<PrinterState>("idle");

  const [currentJob, setCurrentJob] = useState<PrintJob | null>(null);

  const { subscribe } = useWebSocketContext();

  const handleStarted = useCallback((message: WebSocketMessage) => {
    if (message.type !== "printer.started") {
      return;
    }

    const job = message.data as PrintJob;

    setState("printing");
    setCurrentJob(job);
  }, []);

  const handleCompleted = useCallback((message: WebSocketMessage) => {
    if (message.type !== "printer.completed") {
      return;
    }

    const job = message.data as PrintJob;

    setState("completed");
    setCurrentJob(job);
  }, []);

  useEffect(() => {
    const unsubscribeStarted = subscribe("printer.started", handleStarted);

    const unsubscribeCompleted = subscribe(
      "printer.completed",
      handleCompleted,
    );

    return () => {
      unsubscribeStarted();
      unsubscribeCompleted();
    };
  }, [subscribe, handleStarted, handleCompleted]);

  return {
    state,
    currentJob,
  };
}
