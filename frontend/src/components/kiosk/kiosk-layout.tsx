import type { PrinterState } from "../../hooks/use-printer";
import type { PrintJob } from "../../types/printer";
import type { ScreenMode } from "../../types/screen";
import type { WebSocketStatus } from "../../types/websocket";
import type { Notification } from "../../types/notification";
import { HomeLayout } from "../layouts/home-layout";
import { NotificationLayout } from "../layouts/notification-layout";
import type { Activity } from "../widgets/activity-widget";

type KioskLayoutProps = {
  mode: ScreenMode;
  connectionStatus: WebSocketStatus;
  printerState: PrinterState;
  currentJob: PrintJob | null;
  notifications: Notification[];
  activities: Activity[];
};

export function KioskLayout({
  mode,
  connectionStatus,
  printerState,
  currentJob,
  notifications,
  activities,
}: KioskLayoutProps) {
  if (mode === "notification") {
    return (
      <NotificationLayout
        connectionStatus={connectionStatus}
        notifications={notifications}
      />
    );
  }

  return (
    <HomeLayout
      connectionStatus={connectionStatus}
      printerState={printerState}
      currentJob={currentJob}
      activities={activities}
    />
  );
}
