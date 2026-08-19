import {
  Info,
  Monitor,
  Printer,
  Wifi,
  type LucideIcon,
} from "lucide-react";

import type { Activity } from "../../types/activity";
import type { NotificationContext } from "../../types/notification";

type ActivityWidgetProps = {
  activities: Activity[];
};

const activityIcons: Record<NotificationContext, LucideIcon> = {
  printer: Printer,
  system: Info,
  network: Wifi,
  display: Monitor,
};

function formatTimestamp(timestamp: string) {
  const date = new Date(timestamp);

  if (Number.isNaN(date.getTime())) {
    return null;
  }

  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function ActivityWidget({ activities }: ActivityWidgetProps) {
  return (
    <section className="widget activity-widget">
      <div className="widget-title">
        <span>Recent activity</span>
      </div>

      {activities.length === 0 ? (
        <p className="activity-widget__empty">No recent activity</p>
      ) : (
        <ul className="activity-widget__list">
          {activities.map((activity) => (
            <li
              className="activity-widget__item"
              key={activity.id}
            >
              <ActivityIcon context={activity.context} />

              <div className="activity-widget__content">
                <strong className="activity-widget__title">{activity.title}</strong>

                {activity.description && (
                  <p className="activity-widget__message">
                    {activity.description}
                  </p>
                )}
              </div>

              <time
                className="activity-widget__time"
                dateTime={activity.timestamp}
              >
                {formatTimestamp(activity.timestamp)}
              </time>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function ActivityIcon({ context }: { context: NotificationContext }) {
  const Icon = activityIcons[context];

  return (
    <span className="activity-widget__icon" aria-hidden="true">
      <Icon size={18} />
    </span>
  );
}
