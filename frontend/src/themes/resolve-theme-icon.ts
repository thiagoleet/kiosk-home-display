import { defaultIcons } from "./icons/default-icons";
import type { ThemeIconName, ThemeIconProps } from "../types/theme";
import type { ComponentType } from "react";

export function resolveThemeIcon(
  name: ThemeIconName,
  themeIcons?: Partial<Record<ThemeIconName, ComponentType<ThemeIconProps>>>,
) {
  return themeIcons?.[name] ?? defaultIcons[name];
}
