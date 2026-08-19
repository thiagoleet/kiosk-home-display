import { useCallback, useEffect, useState } from "react";

import type { AppState } from "../types/app-state";
import type { DisplayState } from "../types/display";
import type { WebSocketMessage } from "../types/websocket";
import { useWebSocketContext } from "./use-websocket-context";

const initialState: AppState = {
  display: {
    power: "on",
    brightness: 100,
  },
};

export function useKioskState() {
  const [state, setState] = useState<AppState>(initialState);

  const { subscribe } = useWebSocketContext();

  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case "state.snapshot":
        setState(message.data as AppState);
        break;

      case "display.state_changed":
        setState((current) => ({
          ...current,
          display: message.data as DisplayState,
        }));
        break;
    }
  }, []);

  useEffect(() => {
    const unsubscribeSnapshot = subscribe("state.snapshot", handleMessage);

    const unsubscribeDisplay = subscribe(
      "display.state_changed",
      handleMessage,
    );

    return () => {
      unsubscribeSnapshot();
      unsubscribeDisplay();
    };
  }, [subscribe, handleMessage]);

  return {
    state,
  };
}
