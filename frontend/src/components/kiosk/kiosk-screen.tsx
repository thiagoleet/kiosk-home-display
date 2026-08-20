import { useMemo, type ReactNode } from "react";

import type { ScreenMode } from "../../types/screen";
import { useKioskState } from "../../hooks/use-kiosk-state";
import { ScreenSaver } from "../screensaver/screensaver";

type KioskScreenProps = {
  mode: ScreenMode;
  children: ReactNode;
};

export function KioskScreen({ mode, children }: KioskScreenProps) {
  const { state } = useKioskState();

  const isScreenOn = useMemo<boolean>(
    () => state.display.power === "on",
    [state.display.power],
  );

  const ScreenDisplay = useMemo(() => {
    if (isScreenOn) {
      return children;
    }

    return <ScreenSaver />;
  }, [children, isScreenOn]);

  return (
    <main
      className="kiosk-screen"
      data-screen-mode={mode}
      data-screen-on={isScreenOn}
    >
      {ScreenDisplay}
    </main>
  );
}
