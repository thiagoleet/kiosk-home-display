import { useWeatherWidget } from "@/hooks/use-weather-widget";
import { ThemeIcon } from "../theme/theme-icon";

export function WeatherViewWidget() {
  const { icon, temperature } = useWeatherWidget();

  return (
    <section className="kiosk-weather">
      {icon && (
        <ThemeIcon
          name={icon}
          size={96}
        />
      )}
      <span>{temperature}</span>
    </section>
  );
}
