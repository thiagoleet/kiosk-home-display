import { defaultIcons } from "./icons/default-icons";

import type {
  ThemeIconComponent,
  ThemeIconName,
  ThemeIconRegistry,
} from "./theme-icons";

export function resolveThemeIcon(
  name: ThemeIconName,
  themeIcons?: ThemeIconRegistry,
): ThemeIconComponent {
  return themeIcons?.[name] ?? defaultIcons[name]!;
}
