import { useWeatherWidget } from "@/hooks/use-weather-widget";
import { ThemeIcon } from "../theme/theme-icon";

export function KioskWeather() {
  const { icon, temperature } = useWeatherWidget();

  return (
    <section className="kiosk-weather">
      {icon && <ThemeIcon name={icon} />}
      <span>{temperature}</span>
    </section>
  );
}
