import { Notification } from "./notification";

import type { Notification as NotificationData } from "@/types/notification";

type NotificationListProps = {
  notification: NotificationData | null;
};

export function NotificationList({ notification }: NotificationListProps) {
  if (!notification) {
    return null;
  }

  return (
    <div className="notification-list">
      <Notification notification={notification} />
    </div>
  );
}
