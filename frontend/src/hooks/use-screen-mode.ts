import { useMemo } from "react";

import type { Notification } from "../types/notification";
import type { ScreenMode } from "../types/screen";

type UseScreenModeOptions = {
  notification: Notification | null;
  isScreenOn: boolean;
};

export function useScreenMode({
  notification,
  isScreenOn,
}: UseScreenModeOptions) {
  const mode = useMemo<ScreenMode>(() => {
    if (!isScreenOn) {
      return "screensaver";
    }

    return notification ? "notification" : "home";
  }, [notification, isScreenOn]);

  return {
    mode,
  };
}
