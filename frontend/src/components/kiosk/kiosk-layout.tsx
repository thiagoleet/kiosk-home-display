import { KioskHeader } from "./kiosk-header";

import { HomeLayout } from "../layouts/home-layout";
import { NotificationLayout } from "../layouts/notification-layout";

import type { Notification } from "../../types/notification";
import type { ScreenMode } from "../../types/screen";
import type { WebSocketStatus } from "../../hooks/use-websocket";
import type { PrinterState } from "../../hooks/use-printer";
import type { PrintJob } from "../../types/printer";
import type { Activity } from "../widgets/activity-widget";

type KioskLayoutProps = {
  mode: ScreenMode;
  connectionStatus: WebSocketStatus;
  printerState: PrinterState;
  currentJob: PrintJob | null;
  notification: Notification | null;
  activities: Activity[];
};

export function KioskLayout({
  mode,
  connectionStatus,
  printerState,
  currentJob,
  notification,
  activities,
}: KioskLayoutProps) {
  return (
    <section className="kiosk-layout">
      <KioskHeader
        status={connectionStatus}
        hasNotification={notification !== null}
      />

      <div
        className="kiosk-layout__content"
        data-mode={mode}
      >
        {mode === "notification" && notification ? (
          <NotificationLayout notification={notification} />
        ) : (
          <HomeLayout
            printerState={printerState}
            currentJob={currentJob}
            activities={activities}
          />
        )}
      </div>
    </section>
  );
}
