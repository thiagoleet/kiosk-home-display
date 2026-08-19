import type { ReactNode } from "react";

import type { ScreenMode } from "../../types/screen";

type KioskScreenProps = {
  mode: ScreenMode;
  children: ReactNode;
};

export function KioskScreen({ mode, children }: KioskScreenProps) {
  return (
    <main
      className="kiosk-screen"
      data-screen-mode={mode}
    >
      {children}
    </main>
  );
}
