import { useCallback, useEffect, useState } from "react";

import type { DisplayState } from "../types/display";
import type { WebSocketMessage } from "../types/websocket";
import { useWebSocketContext } from "./use-websocket-context";

// The daemon announces the idle timeout while the screen is still lit and only
// powers it down after its sleep delay, so the screensaver has to come up on
// the event instead of waiting for the display to report itself off — by then
// there is nothing left to look at.
export function useIdle() {
  const [isIdle, setIsIdle] = useState(false);

  const { subscribe } = useWebSocketContext();

  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case "idle.timeout":
        setIsIdle(true);
        break;

      // Both cases below mirror what restarts the daemon's countdown: a wake,
      // and a notification, which the daemon also treats as activity. Without
      // the notification case the screensaver would come back when the
      // notification expires, on a screen the daemon no longer considers idle.
      case "display.state_changed":
        if ((message.data as DisplayState).power === "on") {
          setIsIdle(false);
        }
        break;

      case "notification":
        setIsIdle(false);
        break;
    }
  }, []);

  useEffect(() => {
    const unsubscribeTimeout = subscribe("idle.timeout", handleMessage);

    const unsubscribeDisplay = subscribe(
      "display.state_changed",
      handleMessage,
    );

    const unsubscribeNotification = subscribe("notification", handleMessage);

    return () => {
      unsubscribeTimeout();
      unsubscribeDisplay();
      unsubscribeNotification();
    };
  }, [subscribe, handleMessage]);

  return {
    isIdle,
  };
}
