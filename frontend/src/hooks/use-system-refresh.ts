import { useCallback, useEffect } from "react";

import { useWebSocketContext } from "./use-websocket-context";

export function useSystemRefresh() {
  const { subscribe } = useWebSocketContext();

  const handleMessage = useCallback(() => {
    window.location.reload();
  }, []);

  useEffect(() => {
    return subscribe("system.refresh", handleMessage);
  }, [subscribe, handleMessage]);
}
