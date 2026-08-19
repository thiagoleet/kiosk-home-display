import { useCallback, useEffect, useState } from "react";

import { useWebSocketContext } from "../hooks/use-websocket-context";
import type { Notification } from "../types/notification";
import type { WebSocketMessage } from "../types/websocket";

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const { subscribe } = useWebSocketContext();

  const removeNotification = useCallback((id: string) => {
    setNotifications((current) =>
      current.filter((notification) => notification.id !== id),
    );
  }, []);

  const handleMessage = useCallback(
    (message: WebSocketMessage) => {
      if (message.type !== "notification") {
        return;
      }

      const notification = message.data as Notification;

      setNotifications((current) => [...current, notification]);

      if (notification.duration > 0) {
        window.setTimeout(() => {
          removeNotification(notification.id);
        }, notification.duration);
      }
    },
    [removeNotification],
  );

  useEffect(() => {
    return subscribe("notification", handleMessage);
  }, [subscribe, handleMessage]);

  return {
    notifications,
    removeNotification,
  };
}
