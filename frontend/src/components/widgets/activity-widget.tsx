import { Info, Monitor, Printer, Wifi, type LucideIcon } from "lucide-react";

import type { NotificationContext } from "../../types/notification";

import { useTranslation } from "../../hooks/use-translation";
import { useActivities } from "../../hooks/use-activities";

const activityIcons: Record<NotificationContext, LucideIcon> = {
  printer: Printer,
  system: Info,
  network: Wifi,
  display: Monitor,
};

function formatTimestamp(timestamp: string, locale: string) {
  const date = new Date(timestamp);

  if (Number.isNaN(date.getTime())) {
    return null;
  }

  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function ActivityWidget() {
  const { locale, t } = useTranslation();

  const { activities } = useActivities();

  return (
    <section className="widget activity-widget">
      <div className="widget-title">
        <span>{t("activity.title")}</span>
      </div>

      {activities.length === 0 ? (
        <p className="activity-widget__empty">{t("activity.empty")}</p>
      ) : (
        <ul className="activity-widget__list">
          {activities.map((activity) => (
            <li
              className="activity-widget__item"
              key={activity.id}
            >
              <ActivityIcon context={activity.context} />

              <div className="activity-widget__content">
                <strong className="activity-widget__title">
                  {activity.title}
                </strong>

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
                {formatTimestamp(activity.timestamp, locale)}
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
    <span
      className="activity-widget__icon"
      aria-hidden="true"
    >
      <Icon size={18} />
    </span>
  );
}
