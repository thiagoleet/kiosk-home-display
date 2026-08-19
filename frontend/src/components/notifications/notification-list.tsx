// import { NotificationIcon } from "./notification-icon";

import type { Notification } from "../../types/notification";

type NotificationListProps = {
  notifications: Notification[];
};

export function NotificationList({ notifications }: NotificationListProps) {
  return (
    <section>
      {notifications.map((notification) => (
        <article
          key={notification.id}
          data-level={notification.level}
        >
          {/* <NotificationIcon context={notification.context} /> */}

          <div>
            <strong>{notification.title}</strong>

            <p>{notification.message}</p>
          </div>
        </article>
      ))}
    </section>
  );
}
