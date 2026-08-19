import { PrinterWidget } from "../widgets/printer-widget";
import { ActivityWidget } from "../widgets/activity-widget";

import type { WebSocketStatus } from "../../types/websocket";
import type { PrintJob } from "../../types/printer";
import type { PrinterState } from "../../hooks/use-printer";
import type { Activity } from "../widgets/activity-widget";
import { KioskHeader } from "../kiosk/kiosk-header";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskClock } from "../kiosk/kiosk-clock";

type HomeLayoutProps = {
  kioskName: string;
  connectionStatus: WebSocketStatus;
  printerState: PrinterState;
  currentJob: PrintJob | null;
  activities: Activity[];
};

export function HomeLayout({
  kioskName,
  connectionStatus,
  printerState,
  currentJob,
  activities,
}: HomeLayoutProps) {
  return (
    <section className="home-layout">
      <KioskHeader
        name={kioskName}
        status={connectionStatus}
      />

      <div className="home-layout__hero">
        <KioskGreeting />

        <div className="home-layout__time">
          <KioskClock />
        </div>
      </div>

      <div className="home-layout__widgets">
        <PrinterWidget
          state={printerState}
          job={currentJob}
        />

        <ActivityWidget activities={activities} />
      </div>
    </section>
  );
}
