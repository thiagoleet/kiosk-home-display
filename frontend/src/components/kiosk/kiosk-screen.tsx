import { useMemo, type ReactNode } from "react";

import type { ScreenMode } from "../../types/screen";
import { useKioskState } from "../../hooks/use-kiosk-state";
import { ScreenSaver } from "../screensaver/screensaver";
import { KioskTransition } from "./kiosk-transition";

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

  const ScreenDisplay = isScreenOn ? children : <ScreenSaver />;

  return (
    <main
      className="kiosk-screen"
      data-screen-mode={mode}
      data-screen-on={isScreenOn}
    >
      <KioskTransition
        mode="screen"
        transitionKey={isScreenOn ? "on" : "off"}
      >
        {ScreenDisplay}
      </KioskTransition>
    </main>
  );
}
