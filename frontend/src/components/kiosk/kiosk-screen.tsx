import type { ReactNode } from "react";

import { useTheme } from "../../hooks/use-theme";
import type { ScreenMode } from "../../types/screen";

type KioskScreenProps = {
  mode: ScreenMode;
  children: ReactNode;
};

export function KioskScreen({ mode, children }: KioskScreenProps) {
  const theme = useTheme();

  return (
    <main
      className="kiosk-screen"
      data-screen-mode={mode}
      style={{
        backgroundColor: theme.colors.background,
        color: theme.colors.foreground,
        fontFamily: theme.typography.fontFamily,
      }}
    >
      {children}
    </main>
  );
}
