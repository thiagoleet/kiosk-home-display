import { useWeatherWidget } from "@/hooks/use-weather-widget";
import { ThemeIcon } from "../theme/theme-icon";
import { useTranslation } from "@/hooks/use-translation";

export function WeatherViewWidget() {
  const { icon, temperature } = useWeatherWidget();

  return (
    <section className="kiosk-weather">
      {icon && (
        <ThemeIcon
          name={icon}
          size={48}
        />
      )}
      <span>{temperature}</span>
    </section>
  );
}

export function WeatherViewWidgetFull() {
  const { icon, temperature, weather } = useWeatherWidget();
  const { t } = useTranslation();

  const conditionLabel = weather?.condition
    ? t(`weather.condition.${weather.condition}`)
    : "—";

  return (
    <section className="kiosk-weather-wrapper">
      <section className="kiosk-weather">
        {icon && (
          <ThemeIcon
            name={icon}
            size={96}
          />
        )}
        <span>{temperature}</span>
      </section>

      <section className="kiosk-weather-details">
        <span>
          {t("weather.condition", {
            condition: conditionLabel,
          })}
        </span>
        <span>
          {t("weather.humidity", {
            humidity: weather?.humidity ?? "—",
          })}
        </span>
        <span>
          {t("weather.windSpeed", {
            windSpeed: weather?.windSpeed ?? "—",
          })}
        </span>
      </section>
    </section>
  );
}
