import { createContext } from "react";

import type { KioskTheme } from "../types/theme";

export type ThemeContextValue = {
  theme: KioskTheme;
};

export const ThemeContext = createContext<ThemeContextValue | null>(null);
