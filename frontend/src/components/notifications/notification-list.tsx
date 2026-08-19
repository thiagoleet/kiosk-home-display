import { Notification } from "./notification";

import type { Notification as NotificationData } from "../../types/notification";

type NotificationListProps = {
  notifications: NotificationData[];
};

const MAX_VISIBLE_NOTIFICATIONS = 3;

export function NotificationList({ notifications }: NotificationListProps) {
  const visibleNotifications = notifications.slice(-MAX_VISIBLE_NOTIFICATIONS);

  return (
    <div className="notification-list">
      {visibleNotifications.map((notification) => (
        <Notification
          key={notification.id}
          notification={notification}
        />
      ))}
    </div>
  );
}
