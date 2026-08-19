import type { KioskTheme } from "../../types/theme";

export const modernTheme: KioskTheme = {
  id: "modern",

  colors: {
    background: "#0f1115",
    foreground: "#f1f5f9",
    primary: "#60a5fa",
    secondary: "#94a3b8",
    muted: "#64748b",
    border: "#334155",

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
};
