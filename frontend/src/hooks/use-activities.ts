import { useCallback, useEffect, useState } from "react";

import { getActivities } from "../services/activity-service";

import { useWebSocketContext } from "./use-websocket-context";

import { MAX_ACTIVITIES } from "../constants/activity";
import type { Activity } from "../types/activity";
import type { WebSocketMessage } from "../types/websocket";

type UseActivitiesProps = {
  maxActivities?: number;
};

export function useActivities({
  maxActivities = MAX_ACTIVITIES,
}: UseActivitiesProps) {
  const [activities, setActivities] = useState<Activity[]>([]);

  const { subscribe } = useWebSocketContext();

  const addActivity = useCallback(
    (activity: Activity) => {
      setActivities((current) =>
        [activity, ...current].slice(0, maxActivities),
      );
    },
    [maxActivities],
  );

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
    let cancelled = false;

    getActivities()
      .then((activities) => {
        if (cancelled) {
          return;
        }

        setActivities(activities.slice(0, maxActivities));
      })
      .catch((error) => {
        console.error("Failed to load activities:", error);
      });

    return () => {
      cancelled = true;
    };
  }, [maxActivities]);

  useEffect(() => {
    return subscribe("activity", handleMessage);
  }, [subscribe, handleMessage]);

  return {
    activities,
  };
}
