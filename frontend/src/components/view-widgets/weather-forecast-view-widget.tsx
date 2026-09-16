import { useWeatherForecastWidget } from "@/hooks/use-weather-forecast-widget";
import type { WeatherForecast } from "@/types/weather";

import { formatDate } from "date-fns";
import { ptBR, enUS } from "date-fns/locale";
import { ThemeIcon } from "../theme/theme-icon";
import { useTranslation } from "@/hooks/use-translation";

type ForecastItemProps = Readonly<{
  item: WeatherForecast;
  index: number;
}>;

function ForecastItem({ item, index }: ForecastItemProps) {
  const { t, locale } = useTranslation();

  const formattedDate = formatDate(new Date(item.date), "EE", {
    locale: locale === "pt-BR" ? ptBR : enUS,
  });

  const widgetTitle = index === 0 ? t("forecast.today") : formattedDate;

  return (
    <div className="widget forecast-widget__item">
      <span className="widget-title ">{widgetTitle}</span>
      {item.icon && (
        <ThemeIcon
          name={item.icon}
          size={128}
        />
      )}

      <span className="widget-text">
        {t("forecast.min")}: {item.temperatureMin}
      </span>

      <span className="widget-text">
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
      <header>
        <h2>{t("forecast.title")}</h2>
      </header>

      <div className="forecast-view__content">
        {patchetForecast.map((item, index) => (
          <ForecastItem
            key={item.date}
            item={item}
            index={index}
          />
        ))}
      </div>
    </section>
  );
}
