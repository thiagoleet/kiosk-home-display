import { NotificationIcon } from "./notification-icon";

import type { Notification as NotificationData } from "../../types/notification";

type NotificationProps = {
  notification: NotificationData;
};

export function Notification({ notification }: NotificationProps) {
  return (
    <article
      className="notification"
      data-context={notification.context}
      data-level={notification.level}
    >
      <div className="notification__icon">
        <NotificationIcon context={notification.context} />
      </div>

      <div className="notification__content">
        <strong className="notification__title">{notification.title}</strong>

        <p className="notification__message">{notification.message}</p>
      </div>
    </article>
  );
}
