import { WeatherIcon } from "../weather/weather-icon";

export function KioskWeather() {
  return (
    <section className="kiosk-weather">
      <WeatherIcon /> <span>32ºC</span>
    </section>
  );
}
