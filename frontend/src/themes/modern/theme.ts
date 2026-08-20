import notificationSound from "../../assets/sounds/notification.mp3";
import type { KioskTheme } from "@/types/theme";

export const modernTheme: KioskTheme = {
  id: "modern",

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
    fontFamily: '"Inter", "Segoe UI", sans-serif',

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
  icons: {},
};
