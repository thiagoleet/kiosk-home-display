import {
  Bell,
  CheckCircle,
  Circle,
  Cloud,
  CloudDrizzle,
  CloudFog,
  CloudLightning,
  CloudRain,
  CloudSnow,
  CloudSun,
  LoaderCircle,
  Printer,
  Sun,
  Wifi,
  WifiOff,
} from "lucide-react";

import type { KioskThemeIcons } from "@/types/theme";

export const defaultIcons: KioskThemeIcons = {
  notification: Bell,
  loading: LoaderCircle,

  "notification.info": Bell,
  "notification.success": CheckCircle,
  "notification.warning": Bell,
  "notification.error": Bell,

  printer: Printer,
  status: Circle,
  "status.online": Circle,
  "status.offline": Circle,

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
