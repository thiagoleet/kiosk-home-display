export type KioskThemeId = "modern" | "retro-16bit";

export type KioskTheme = {
  id: KioskThemeId;

  colors: {
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

  typography: {
    fontFamily: string;
    headingWeight: number;
    bodyWeight: number;
  };

  spacing: {
    xs: string;
    sm: string;
    md: string;
    lg: string;
    xl: string;
  };

  border: {
    width: string;
    radius: string;
  };
};
