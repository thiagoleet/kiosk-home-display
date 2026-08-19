import { useCallback, useEffect, useState } from "react";

import { useWebSocketContext } from "../hooks/use-websocket-context";
import type { Notification } from "../types/notification";
import type { WebSocketMessage } from "../types/websocket";

type NotificationState = {
  active: Notification | null;
  queue: Notification[];
};

export function useNotifications() {
  const [state, setState] = useState<NotificationState>({
    active: null,
    queue: [],
  });

  const { subscribe } = useWebSocketContext();

  const showNextNotification = useCallback(() => {
    setState((current) => {
      const [next, ...remaining] = current.queue;

      return {
        active: next ?? null,
        queue: remaining,
      };
    });
  }, []);

  const removeNotification = useCallback((id: string) => {
    setState((current) => {
      if (current.active?.id === id) {
        const [next, ...remaining] = current.queue;

        return {
          active: next ?? null,
          queue: remaining,
        };
      }

      return {
        ...current,
        queue: current.queue.filter((notification) => notification.id !== id),
      };
    });
  }, []);

  const handleMessage = useCallback((message: WebSocketMessage) => {
    if (message.type !== "notification") {
      return;
    }

    const notification = message.data as Notification;

    setState((current) => {
      if (current.active) {
        return {
          active: current.active,
          queue: [...current.queue, notification],
        };
      }

      return {
        active: notification,
        queue: current.queue,
      };
    });
  }, []);

  useEffect(() => {
    return subscribe("notification", handleMessage);
  }, [subscribe, handleMessage]);

  useEffect(() => {
    const notification = state.active;

    if (!notification || notification.duration <= 0) {
      return;
    }

    const timer = window.setTimeout(() => {
      removeNotification(notification.id);
    }, notification.duration);

    return () => {
      window.clearTimeout(timer);
    };
  }, [state.active, removeNotification]);

  return {
    activeNotification: state.active,
    notificationQueue: state.queue,
    removeNotification,
    showNextNotification,
  };
}
