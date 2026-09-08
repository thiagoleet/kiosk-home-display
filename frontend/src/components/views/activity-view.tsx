import { ActivityWidget } from "../widgets/activity-widget";

export function ActivityView() {
  return (
    <div className="activity-view">
      <ActivityWidget maxActivities={8} />
    </div>
  );
}
