import type { ComponentType } from "react";

export type ThemeIconName =
  | "notification"
  | "notification.info"
  | "notification.success"
  | "notification.warning"
  | "notification.error"
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
  | "weather.thunderstorm";

export type ThemeIconProps = {
  size?: number;
  className?: string;
};

export type ThemeIconComponent = ComponentType<ThemeIconProps>;

export type ThemeIconRegistry = Partial<
  Record<ThemeIconName, ThemeIconComponent>
>;
