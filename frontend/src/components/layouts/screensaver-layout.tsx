import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskWeather } from "../kiosk/kiosk-weather";

export function ScreenSaverLayout() {
  return (
    <section className="screensaver-layout">
      <div className="screensaver-layout__main">
        <div className="screensaver-layout__greeting">
          <KioskGreeting />
          <KioskClock />
          <KioskWeather />
        </div>
      </div>
    </section>
  );
}
