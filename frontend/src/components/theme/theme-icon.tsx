import type { ComponentProps } from "react";

import { resolveThemeIcon } from "../../theme/resolve-theme-icon";
import type { ThemeIconName } from "../../theme/theme-icons";
import { useTheme } from "../../hooks/use-theme";

type ThemeIconProps = {
  name: ThemeIconName;
} & Omit<ComponentProps<"svg">, "name">;

export function ThemeIcon({ name, ...props }: ThemeIconProps) {
  const { icons } = useTheme();

  const Icon = resolveThemeIcon(name, icons);

  return <Icon {...props} />;
}
