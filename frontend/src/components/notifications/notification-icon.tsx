import { Bell, Network, Printer, Settings } from "lucide-react";

import type { NotificationContext } from "../../types/notification";

type NotificationIconProps = {
  context: NotificationContext;
};

export function NotificationIcon({ context }: NotificationIconProps) {
  switch (context) {
    case "printer":
      return <Printer />;

    case "network":
      return <Network />;

    case "system":
      return <Settings />;

    default:
      return <Bell />;
  }
}
