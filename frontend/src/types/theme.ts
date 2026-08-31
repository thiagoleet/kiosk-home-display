import type { ComponentType } from "react";

export type KioskThemeId = "modern" | "retro-16bit";

export type KioskTheme = {
  id: KioskThemeId;
  colors: KioskThemeColors;
  typography: KioskThemeTypography;
  spacing: KioskThemeSpacing;
  border: KioskThemeBorder;
  sounds: KioskThemeSounds;
  icons: KioskThemeIcons;
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

export type KioskThemeIcons = Partial<
  Record<ThemeIconName, ComponentType<ThemeIconProps>>
>;

export type ThemeIconName =
  | "notification"
  | "notification.info"
  | "notification.success"
  | "notification.warning"
  | "notification.error"
  | "loading"
  | "printer"
  | "wifi"
  | "wifi-off"
  | "weather.clear"
  | "weather.partly_cloudy"
  | "weather.overcast"
  | "weather.fog"
  | "weather.drizzle"
  | "weather.rain"
  | "weather.rain_showers"
  | "weather.snow"
  | "weather.thunderstorm"
  | "status"
  | "status.offline"
  | "status.online";

export type ThemeIconProps = {
  size?: number;
  className?: string;
};
