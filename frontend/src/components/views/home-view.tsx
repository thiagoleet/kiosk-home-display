import { KioskViewWidget } from "../view-widgets/clock-view-widget";
import { GreetingViewWidget } from "../view-widgets/greeting-view-widget";
import { KioskWeather } from "../view-widgets/weather-view-widget";

export function HomeView() {
  return (
    <div className="home-layout__greeting">
      <GreetingViewWidget />
      <KioskViewWidget />
      <KioskWeather />
    </div>
  );
}
