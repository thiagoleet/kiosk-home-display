import { useCallback, useMemo, type PropsWithChildren } from "react";

import { useWebSocket } from "../hooks/use-websocket";
import { websocketUrl } from "../config/websocket-config";
import {
  WebSocketContext,
  type WebSocketContextValue,
} from "../hooks/use-websocket-context";
import { WebSocketDispatcher } from "../services/websocket-dispatcher";

export function WebSocketProvider({ children }: PropsWithChildren) {
  const dispatcher = useMemo(() => new WebSocketDispatcher(), []);

  const handleMessage = useCallback(
    (message: Parameters<typeof dispatcher.dispatch>[0]) => {
      dispatcher.dispatch(message);
    },
    [dispatcher],
  );

  const { status } = useWebSocket({
    url: websocketUrl,
    onMessage: handleMessage,
  });

  const subscribe = dispatcher.subscribe.bind(dispatcher);

  const value = useMemo<WebSocketContextValue>(
    () => ({
      status,
      subscribe,
    }),
    [status, subscribe],
  );

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}
