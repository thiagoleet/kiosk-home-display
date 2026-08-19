import { useCallback, useEffect, useRef, useState } from "react";

import { useTheme } from "./use-theme";
import { useWebSocketContext } from "./use-websocket-context";
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
  const notificationAudioRef = useRef<HTMLAudioElement | null>(null);

  const theme = useTheme();
  const { subscribe } = useWebSocketContext();

  useEffect(() => {
    const audio = new Audio(theme.sounds.notification);

    audio.preload = "auto";
    notificationAudioRef.current = audio;

    return () => {
      audio.pause();
      notificationAudioRef.current = null;
    };
  }, [theme.sounds.notification]);

  const playNotificationSound = useCallback(() => {
    const audio = notificationAudioRef.current;

    if (!audio) {
      return;
    }

    audio.currentTime = 0;
    void audio.play().catch(() => {
      // Browsers can block audio until the user has interacted with the page.
    });
  }, []);

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

    playNotificationSound();

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
  }, [playNotificationSound]);

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
