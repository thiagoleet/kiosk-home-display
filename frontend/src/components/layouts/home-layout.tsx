import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskWeather } from "../kiosk/kiosk-weather";

import { ActivityWidget } from "../widgets/activity-widget";
import { PrinterWidget } from "../widgets/printer-widget";

export function HomeLayout() {
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
        <PrinterWidget />
        <ActivityWidget />
      </div>
    </section>
  );
}
