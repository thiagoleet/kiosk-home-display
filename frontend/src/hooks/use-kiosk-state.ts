import { useCallback, useState } from "react";

import type { AppState } from "../types/app-state";
import type { DisplayState } from "../types/display";
import type { WebSocketMessage } from "../types/websocket";

const initialState: AppState = {
  display: {
    power: "on",
    brightness: 100,
  },
};

export function useKioskState() {
  const [state, setState] = useState<AppState>(initialState);

  const handleMessage = useCallback((message: WebSocketMessage) => {
    console.log("[WebSocket]", message);

    switch (message.type) {
      case "state.snapshot":
        console.log("[WebSocket] snapshot", message.data);

        setState(message.data as AppState);
        break;

      case "display.state_changed":
        console.log("[WebSocket] display changed", message.data);

        setState((current) => ({
          ...current,
          display: message.data as DisplayState,
        }));

        break;
    }
  }, []);

  return {
    state,
    handleMessage,
  };
}
