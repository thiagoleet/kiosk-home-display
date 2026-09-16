import { useMemo } from "react";

import { screensaverEnabled } from "../config/screensaver-config";
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
    if (screensaverEnabled && !isScreenOn) {
      return "screensaver";
    }

    // A notification outranks the idle screensaver: the daemon wakes the
    // display and restarts its countdown for one, so the screensaver would be
    // covering a screen that is no longer idle.
    if (notification) {
      return "notification";
    }

    return screensaverEnabled && isIdle ? "screensaver" : "home";
  }, [notification, isScreenOn, isIdle]);

  return {
    mode,
  };
}
