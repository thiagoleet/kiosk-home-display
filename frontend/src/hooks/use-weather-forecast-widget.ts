import type { WeatherForecast } from "@/types/weather";
import { useWeatherForecast } from "./use-weather-forecast";

export function useWeatherForecastWidget() {
  const { data, isLoading, error, refetch } = useWeatherForecast();

  if (!data) {
    return {
      forecast: null,
      isLoading,
      error,
      refetch,
    };
  }

  const forecast: WeatherForecast[] = data.map((item) => ({
    ...item,
    temperatureMax: `${Math.round(item.temperatureMax)}ºC`,
    temperatureMin: `${Math.round(item.temperatureMin)}ºC`,
    icon: `weather.${item.condition}` as const,
  }));

  return {
    forecast,
    isLoading,
    error,
    refetch,
  };
}
