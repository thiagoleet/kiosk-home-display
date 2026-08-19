import { useMemo } from "react";

import type { Notification } from "../types/notification";
import type { ScreenMode } from "../types/screen";

type UseScreenModeOptions = {
  notifications: Notification[];
};

export function useScreenMode({ notifications }: UseScreenModeOptions) {
  const mode = useMemo<ScreenMode>(() => {
    if (notifications.length > 0) {
      return "notification";
    }

    return "home";
  }, [notifications]);

  return {
    mode,
  };
}
