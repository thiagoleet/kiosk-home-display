import type { Notification } from "../types/notification";

type NotificationListProps = {
  notifications: Notification[];
};

export function NotificationList({ notifications }: NotificationListProps) {
  return (
    <section>
      {notifications.map((notification, index) => (
        <article key={index}>
          <strong>{notification.title}</strong>

          <p>{notification.message}</p>

          <small>{notification.level}</small>
        </article>
      ))}
    </section>
  );
}
