import { useMemo } from "react";

import type { Notification } from "../types/notification";
import type { ScreenMode } from "../types/screen";

type UseScreenModeOptions = {
  notification: Notification | null;
};

export function useScreenMode({ notification }: UseScreenModeOptions) {
  const mode = useMemo<ScreenMode>(
    () => (notification ? "notification" : "home"),
    [notification],
  );

  return {
    mode,
  };
}
