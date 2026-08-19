import type { Notification } from "../types/notification";

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
          <strong>{notification.title}</strong>

          <p>{notification.message}</p>
        </article>
      ))}
    </section>
  );
}
