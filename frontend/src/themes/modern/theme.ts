import type { KioskTheme } from "../../types/theme";

export const modernTheme: KioskTheme = {
  id: "modern",

  colors: {
    background: "#ffffff",
    foreground: "#1f2937",
    primary: "#3979e8",
    secondary: "#64748b",
    muted: "#94a3b8",
    border: "#3979e8",

    success: "#16a34a",
    warning: "#d97706",
    error: "#dc2626",
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
