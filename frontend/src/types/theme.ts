export type KioskThemeId = "modern" | "retro-16bit";

export type KioskTheme = {
  id: KioskThemeId;
  colors: KioskThemeColors;
  typography: KioskThemeTypography;
  spacing: KioskThemeSpacing;
  border: KioskThemeBorder;
  sounds: KioskThemeSounds;
};

export type KioskThemeColors = {
  background: string;
  surface: string;
  surfaceElevated: string;

  foreground: string;
  secondary: string;
  muted: string;

  primary: string;
  border: string;

  success: string;
  warning: string;
  error: string;
};

export type KioskThemeTypography = {
  fontFamily: string;
  headingWeight: number;
  bodyWeight: number;
};

export type KioskThemeSpacing = {
  xs: string;
  sm: string;
  md: string;
  lg: string;
  xl: string;
};

export type KioskThemeBorder = {
  width: string;
  radius: string;
};

export type KioskThemeSounds = {
  notification: string;
};
