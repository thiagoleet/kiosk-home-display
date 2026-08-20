import {
  Bell,
  CheckCircle,
  Cloud,
  CloudDrizzle,
  CloudFog,
  CloudLightning,
  CloudRain,
  CloudSnow,
  CloudSun,
  Printer,
  Sun,
  Wifi,
  WifiOff,
} from "lucide-react";

import type { ThemeIconRegistry } from "../theme-icons";

export const defaultIcons: ThemeIconRegistry = {
  notification: Bell,

  "notification.info": Bell,
  "notification.success": CheckCircle,
  "notification.warning": Bell,
  "notification.error": Bell,

  printer: Printer,

  wifi: Wifi,
  "wifi-off": WifiOff,

  "weather.clear": Sun,
  "weather.partly_cloudy": CloudSun,
  "weather.overcast": Cloud,
  "weather.fog": CloudFog,
  "weather.drizzle": CloudDrizzle,
  "weather.rain": CloudRain,
  "weather.rain_showers": CloudRain,
  "weather.snow": CloudSnow,
  "weather.thunderstorm": CloudLightning,
};
