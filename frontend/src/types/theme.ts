export type KioskThemeId = "modern" | "retro-16bit";

export type KioskTheme = {
  id: KioskThemeId;

  colors: {
    background: string;
    foreground: string;
    primary: string;
    secondary: string;
    muted: string;
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
