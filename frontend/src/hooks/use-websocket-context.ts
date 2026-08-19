import { createContext, useContext } from "react";
import type { WebSocketDispatcher } from "../services/websocket-dispatcher";

export type WebSocketStatus = "connecting" | "connected" | "disconnected";

export type WebSocketContextValue = {
  status: WebSocketStatus;
  subscribe: WebSocketDispatcher["subscribe"];
};

export const WebSocketContext = createContext<WebSocketContextValue | null>(
  null,
);

export function useWebSocketContext() {
  const context = useContext(WebSocketContext);

  if (!context) {
    throw new Error(
      "useWebSocketContext must be used inside WebSocketProvider",
    );
  }

  return context;
}
