import type { KioskTheme } from "../types/theme";

export function applyTheme(theme: KioskTheme) {
  const root = document.documentElement;

  root.style.setProperty("--color-background", theme.colors.background);

  root.style.setProperty("--color-foreground", theme.colors.foreground);

  root.style.setProperty("--color-primary", theme.colors.primary);

  root.style.setProperty("--color-secondary", theme.colors.secondary);

  root.style.setProperty("--color-muted", theme.colors.muted);

  root.style.setProperty("--color-border", theme.colors.border);

  root.style.setProperty("--color-success", theme.colors.success);

  root.style.setProperty("--color-warning", theme.colors.warning);

  root.style.setProperty("--color-error", theme.colors.error);

  root.style.setProperty("--font-family", theme.typography.fontFamily);

  root.style.setProperty(
    "--font-heading-weight",
    String(theme.typography.headingWeight),
  );

  root.style.setProperty(
    "--font-body-weight",
    String(theme.typography.bodyWeight),
  );

  root.style.setProperty("--spacing-xs", theme.spacing.xs);

  root.style.setProperty("--spacing-sm", theme.spacing.sm);

  root.style.setProperty("--spacing-md", theme.spacing.md);

  root.style.setProperty("--spacing-lg", theme.spacing.lg);

  root.style.setProperty("--spacing-xl", theme.spacing.xl);

  root.style.setProperty("--border-width", theme.border.width);

  root.style.setProperty("--border-radius", theme.border.radius);
}
