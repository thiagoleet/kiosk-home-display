import { createElement, type ComponentProps } from "react";

import { defaultIcons } from "@/themes/icons/default-icons";
import type { ThemeIconName } from "@/types/theme";
import { useTheme } from "@/hooks/use-theme";

type ThemeIconProps = {
  name: ThemeIconName;
  size?: number;
} & Omit<ComponentProps<"svg">, "name">;

export function ThemeIcon({ name, ...props }: ThemeIconProps) {
  const theme = useTheme();

  const Icon = theme.icons[name] ?? defaultIcons[name];

  if (!Icon) {
    return null;
  }

  return createElement(Icon, props);
}
