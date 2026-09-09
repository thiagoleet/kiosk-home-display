import { useMemo } from "react";

import type { Notification } from "../types/notification";
import type { ScreenMode } from "../types/screen";

type UseScreenModeOptions = {
  notification: Notification | null;
  isScreenOn: boolean;
  isIdle: boolean;
};

export function useScreenMode({
  notification,
  isScreenOn,
  isIdle,
}: UseScreenModeOptions) {
  const mode = useMemo<ScreenMode>(() => {
    if (!isScreenOn) {
      return "screensaver";
    }

    // A notification outranks the idle screensaver: the daemon wakes the
    // display and restarts its countdown for one, so the screensaver would be
    // covering a screen that is no longer idle.
    if (notification) {
      return "notification";
    }

    return isIdle ? "screensaver" : "home";
  }, [notification, isScreenOn, isIdle]);

  return {
    mode,
  };
}
