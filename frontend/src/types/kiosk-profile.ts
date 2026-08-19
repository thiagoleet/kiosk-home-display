import type { KioskThemeId } from "./theme";

export type KioskProfile = {
  id: string;
  name: string;

  features: {
    printer: boolean;
  };

  theme: KioskThemeId;
};
