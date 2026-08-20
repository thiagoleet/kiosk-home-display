import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskWeather } from "../kiosk/kiosk-weather";

import { ActivityWidget } from "../widgets/activity-widget";
import { PrinterWidget } from "../widgets/printer-widget";

import type { PrintJob } from "../../types/printer";
import type { PrinterState } from "../../hooks/use-printer";
import type { Activity } from "../../types/activity";

type HomeLayoutProps = {
  printerState: PrinterState;
  currentJob: PrintJob | null;
  activities: Activity[];
};

export function HomeLayout({
  printerState,
  currentJob,
  activities,
}: HomeLayoutProps) {
  return (
    <section className="home-layout">
      <div className="home-layout__main">
        <div className="home-layout__greeting">
          <KioskGreeting />
          <KioskClock />
          <KioskWeather />
        </div>
      </div>

      <div className="home-layout__footer">
        <PrinterWidget
          state={printerState}
          job={currentJob}
        />

        <ActivityWidget activities={activities} />
      </div>
    </section>
  );
}
