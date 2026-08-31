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
    notification: (props) =>
      createElement(RetroIcon, { icon: Bell, size: props.size }),
    loading: (props) =>
      createElement(AnimatedIcon, { icon: Hourglass, size: props.size }),

    "notification.success": (props) =>
      createElement(RetroIcon, { icon: CheckCircle, size: props.size }),

    "weather.thunderstorm": (props) =>
      createElement(RetroIcon, { icon: Thunder, size: props.size }),
    "weather.clear": (props) =>
      createElement(RetroIcon, { icon: Sun, size: props.size }),
    "weather.partly_cloudy": (props) =>
      createElement(RetroIcon, { icon: CloudSun, size: props.size }),
    "weather.overcast": (props) =>
      createElement(RetroIcon, { icon: Cloud, size: props.size }),
    "weather.fog": (props) =>
      createElement(RetroIcon, { icon: Cloud, size: props.size }),
    "weather.drizzle": (props) =>
      createElement(RetroIcon, { icon: Rain, size: props.size }),
    "weather.rain": (props) =>
      createElement(RetroIcon, { icon: Rain, size: props.size }),
    "weather.rain_showers": (props) =>
      createElement(RetroIcon, { icon: Rain, size: props.size }),
    "weather.snow": (props) =>
      createElement(RetroIcon, { icon: Cloud, size: props.size }),

    printer: (props) =>
      createElement(RetroIcon, { icon: Robot, size: props.size }),

    "status.offline": (props) =>
      createElement(RetroIcon, { icon: Skull, size: props.size }),
    "status.online": (props) =>
      createElement(RetroIcon, { icon: LevelUp, size: props.size }),
  },
};
