import { KioskViewWidget } from "../view-widgets/clock-view-widget";
import { GreetingViewWidget } from "../view-widgets/greeting-view-widget";
import { WeatherViewWidget } from "../view-widgets/weather-view-widget";

export function ScreenSaverLayout() {
  return (
    <section className="screensaver-layout">
      <div className="screensaver-layout__main">
        <div className="screensaver-layout__greeting">
          <GreetingViewWidget />
          <KioskViewWidget />
          <WeatherViewWidget />
        </div>
      </div>
    </section>
  );
}
