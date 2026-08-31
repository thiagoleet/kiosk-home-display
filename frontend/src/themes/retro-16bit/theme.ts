import { createElement } from "react";

import notificationSound from "../../assets/sounds/notification.mp3";
import type { KioskTheme } from "@/types/theme";

import { Cloud, CloudSun, Rain, Sun, Thunder } from "@pxlkit/weather";
import { Robot } from "@pxlkit/ui";

import { AnimatedIcon, RetroIcon } from "../retro/pxlkit-icon";
import { Bell, CheckCircle, Hourglass } from "@pxlkit/feedback";
import { LevelUp, Skull } from "@pxlkit/gamification";

export const retro16bitTheme: KioskTheme = {
  id: "retro-16bit",

  colors: {
    background: "#0b0f14",
    surface: "#111820",
    surfaceElevated: "#18212c",

    foreground: "#f1f5f9",
    secondary: "#94a3b8",
    muted: "#64748b",

    primary: "#60a5fa",
    border: "#263241",

    success: "#4ade80",
    warning: "#fbbf24",
    error: "#f87171",
  },

  typography: {
    fontFamily: '"Press Start 2P", monospace',

    headingWeight: 600,
    bodyWeight: 400,
  },

  spacing: {
    xs: "0.25rem",
    sm: "0.5rem",
    md: "1rem",
    lg: "1.5rem",
    xl: "2rem",
  },

  border: {
    width: "1px",
    radius: "0.75rem",
  },

  sounds: {
    notification: notificationSound,
  },
  icons: {
    notification: () => createElement(RetroIcon, { icon: Bell }),
    loading: () => createElement(AnimatedIcon, { icon: Hourglass }),

    "notification.success": () =>
      createElement(RetroIcon, { icon: CheckCircle }),

    "weather.thunderstorm": () => createElement(RetroIcon, { icon: Thunder }),
    "weather.clear": () => createElement(RetroIcon, { icon: Sun }),
    "weather.partly_cloudy": () => createElement(RetroIcon, { icon: CloudSun }),
    "weather.overcast": () => createElement(RetroIcon, { icon: Cloud }),
    "weather.fog": () => createElement(RetroIcon, { icon: Cloud }),
    "weather.drizzle": () => createElement(RetroIcon, { icon: Rain }),
    "weather.rain": () => createElement(RetroIcon, { icon: Rain }),
    "weather.rain_showers": () => createElement(RetroIcon, { icon: Rain }),
    "weather.snow": () => createElement(RetroIcon, { icon: Cloud }),

    printer: () => createElement(RetroIcon, { icon: Robot }),

    "status.offline": (props) =>
      createElement(RetroIcon, { icon: Skull, ...props }),
    "status.online": (props) =>
      createElement(RetroIcon, { icon: LevelUp, ...props }),
  },
};
