import { Info, Monitor, Printer, Wifi, type LucideIcon } from "lucide-react";

import type { NotificationContext } from "../../types/notification";

type NotificationIconProps = {
  context: NotificationContext;
};

const icons: Record<NotificationContext, LucideIcon> = {
  printer: Printer,
  system: Info,
  network: Wifi,
  display: Monitor,
};

export function NotificationIcon({ context }: NotificationIconProps) {
  const Icon = icons[context];

  return (
    <Icon
      aria-hidden="true"
      size={40}
    />
  );
}
