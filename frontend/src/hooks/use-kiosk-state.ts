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

  return {
    state,
    handleMessage,
  };
}
