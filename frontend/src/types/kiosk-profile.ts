import type { KioskThemeId } from "./theme";
import type { Locale } from "../i18n/translations";

export type KioskProfile = {
  id: string;
  name: string;

  features: {
    printer: boolean;
  };

  theme: KioskThemeId;
  locale: Locale;
};
