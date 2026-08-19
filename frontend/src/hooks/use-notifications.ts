import { useCallback, useEffect, useState } from "react";

import type { Notification } from "../types/notification";
import type { WebSocketMessage } from "../types/websocket";
import { useWebSocketContext } from "./use-websocket-context";

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const { subscribe } = useWebSocketContext();

  const handleMessage = useCallback((message: WebSocketMessage) => {
    if (message.type !== "notification") {
      return;
    }

    setNotifications((current) => [...current, message.data as Notification]);
  }, []);

  useEffect(() => {
    return subscribe("notification", handleMessage);
  }, [subscribe, handleMessage]);

  const removeNotification = useCallback((notification: Notification) => {
    setNotifications((current) =>
      current.filter((item) => item !== notification),
    );
  }, []);

  return {
    notifications,
    removeNotification,
  };
}
