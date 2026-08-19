import type { PropsWithChildren } from "react";

import { useKiosk } from "../hooks/use-kiosk";
import { getTheme } from "../themes";

import { ThemeContext } from "./theme-context";

export function ThemeProvider({ children }: PropsWithChildren) {
  const { profile } = useKiosk();

  const theme = getTheme(profile.theme);

  return (
    <ThemeContext.Provider value={{ theme }}>{children}</ThemeContext.Provider>
  );
}
