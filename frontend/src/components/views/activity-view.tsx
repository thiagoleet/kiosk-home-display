import { ActivityWidget } from "../widgets/activity-widget";

import type { Activity } from "@/types/activity";

type ActivityViewProps = Readonly<{
  activities: Activity[];
}>;

export function ActivityView({ activities }: ActivityViewProps) {
  return (
    <div className="activity-view">
      <ActivityWidget activities={activities} />
    </div>
  );
}
