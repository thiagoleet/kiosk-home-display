import { useWeatherForecastWidget } from "@/hooks/use-weather-forecast-widget";
import type { WeatherForecast } from "@/types/weather";

import { formatDate } from "date-fns";
import { ptBR, enUS } from "date-fns/locale";
import { ThemeIcon } from "../theme/theme-icon";
import { useTranslation } from "@/hooks/use-translation";

type ForecastItemProps = {
  item: WeatherForecast;
};

function ForecastItem({ item }: ForecastItemProps) {
  const { t, locale } = useTranslation();

  const formattedDate = formatDate(new Date(item.date), "EEEE", {
    locale: locale === "pt-BR" ? ptBR : enUS,
  });

  return (
    <div className="widget forecast-widget__item">
      <span>{formattedDate}</span>
      {item.icon && (
        <ThemeIcon
          name={item.icon}
          size={96}
        />
      )}

      <span>
        {t("forecast.min")}: {item.temperatureMin}
      </span>

      <span>
        {t("forecast.max")}: {item.temperatureMax}
      </span>
    </div>
  );
}

export function WeatherForecastViewWidget() {
  const { forecast } = useWeatherForecastWidget();
  const { t } = useTranslation();

  if (!forecast) {
    return <div>{t("forecast.unavailable")}</div>;
  }

  const patchetForecast = forecast.slice(0, 3);

  return (
    <section className="home-view forecast-view">
      <header>{t("forecast.title")}</header>

      <div className="forecast-view__content">
        {patchetForecast.map((item) => (
          <ForecastItem
            key={item.date}
            item={item}
          />
        ))}
      </div>
    </section>
  );
}
