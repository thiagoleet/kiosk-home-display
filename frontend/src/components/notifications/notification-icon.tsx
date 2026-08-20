import { Info, Monitor, Printer, Wifi, type LucideIcon } from "lucide-react";

import type {
  NotificationContext,
  NotificationLevel,
} from "@/types/notification";

type NotificationIconProps = {
  context: NotificationContext;
  level?: NotificationLevel | null;
};

const icons: Record<NotificationContext, LucideIcon> = {
  printer: Printer,
  system: Info,
  network: Wifi,
  display: Monitor,
};

const colors: Record<NotificationLevel, string> = {
  info: "var(--color-primary)",
  success: "var(--color-success)",
  warning: "var(--color-warning)",
  error: "var(--color-error)",
};

export function NotificationIcon({
  context,
  level = null,
}: NotificationIconProps) {
  const Icon = icons[context];
  const color = level ? colors[level] : "var(--color-primary)";

  return (
    <Icon
      aria-hidden="true"
      color={color}
      size={40}
    />
  );
}
