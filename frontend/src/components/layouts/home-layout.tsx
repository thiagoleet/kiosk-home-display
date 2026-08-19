import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskHeader } from "../kiosk/kiosk-header";

import { ActivityWidget } from "../widgets/activity-widget";
import { PrinterWidget } from "../widgets/printer-widget";

import type { WebSocketStatus } from "../../hooks/use-websocket";
import type { PrintJob } from "../../types/printer";
import type { PrinterState } from "../../hooks/use-printer";
import type { Activity } from "../widgets/activity-widget";

type HomeLayoutProps = {
  connectionStatus: WebSocketStatus;
  printerState: PrinterState;
  currentJob: PrintJob | null;
  activities: Activity[];
};

export function HomeLayout({
  connectionStatus,
  printerState,
  currentJob,
  activities,
}: HomeLayoutProps) {
  return (
    <section className="home-layout">
      <KioskHeader status={connectionStatus} />

      <div className="home-layout__main">
        <div className="home-layout__greeting">
          <KioskGreeting />
          <KioskClock />
        </div>
      </div>

      <footer className="home-layout__footer">
        <PrinterWidget
          state={printerState}
          job={currentJob}
        />

        <ActivityWidget activities={activities} />
      </footer>
    </section>
  );
}
