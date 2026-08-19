import { useCallback, useEffect, useState } from "react";

import { useWebSocketContext } from "./use-websocket-context";

import { MAX_ACTIVITIES } from "../constants/activity";
import type { Activity } from "../types/activity";
import type { WebSocketMessage } from "../types/websocket";

export function useActivities() {
  const [activities, setActivities] = useState<Activity[]>([]);

  const { subscribe } = useWebSocketContext();

  const addActivity = useCallback((activity: Activity) => {
    setActivities((current) => [activity, ...current].slice(0, MAX_ACTIVITIES));
  }, []);

  const handleMessage = useCallback(
    (message: WebSocketMessage) => {
      if (message.type !== "activity") {
        return;
      }

      const activity = message.data as Activity;

      addActivity(activity);
    },
    [addActivity],
  );

  useEffect(() => {
    return subscribe("activity", handleMessage);
  }, [subscribe, handleMessage]);

  return {
    activities,
  };
}
