import { useEffect, type PropsWithChildren } from "react";

import { useKiosk } from "../hooks/use-kiosk";
import { getTheme } from "../themes";
import { applyTheme } from "../themes/apply-theme";

import { ThemeContext } from "./theme-context";

export function ThemeProvider({ children }: PropsWithChildren) {
  const { profile } = useKiosk();

  const theme = getTheme(profile.theme);

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme }}>{children}</ThemeContext.Provider>
  );
}
