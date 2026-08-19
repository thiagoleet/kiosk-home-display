import type { KioskTheme, KioskThemeId } from "../types/theme";

import { modernTheme } from "./modern/theme";

const themes: Record<KioskThemeId, KioskTheme> = {
  modern: modernTheme,

  // Temporariamente usamos o modern
  // até implementarmos o tema retro.
  "retro-16bit": modernTheme,
};

export function getTheme(themeId: KioskThemeId): KioskTheme {
  return themes[themeId];
}
