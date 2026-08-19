import type { KioskTheme } from "./theme";

export type KioskProfile = {
  id: string;
  name: string;

  features: {
    printer: boolean;
  };

  theme: KioskTheme;
};
