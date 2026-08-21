import { ClockViewWidget } from "../view-widgets/clock-view-widget";
import { GreetingViewWidget } from "../view-widgets/greeting-view-widget";
import { WeatherViewWidget } from "../view-widgets/weather-view-widget";

export function HomeView() {
  return (
    <div className="home-view">
      <GreetingViewWidget />
      <ClockViewWidget />
      <WeatherViewWidget />
    </div>
  );
}
