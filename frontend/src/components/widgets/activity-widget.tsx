import { Check, Printer } from "lucide-react";

export type Activity = {
  id: string;
  type: "printer";
  message: string;
};

type ActivityWidgetProps = {
  activities: Activity[];
};

export function ActivityWidget({ activities }: ActivityWidgetProps) {
  return (
    <section className="widget activity-widget">
      <div className="widget-title">
        <span>Recent activity</span>
      </div>

      {activities.length === 0 ? (
        <p className="activity-widget__empty">No recent activity</p>
      ) : (
        <div className="activity-widget__list">
          {activities.map((activity) => (
            <div
              className="activity-widget__item"
              key={activity.id}
            >
              <Printer size={18} />

              <span>{activity.message}</span>

              <Check size={16} />
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
