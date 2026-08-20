import { useWeather } from "./use-weather";

import type { ThemeIconName } from "../types/theme";

export function useWeatherWidget() {
  const { data: weather, isLoading, error, refetch } = useWeather();

  if (!weather) {
    return {
      weather: null,
      temperature: null,
      condition: null,
      icon: null,
      isDay: null,
      isLoading,
      error,
      refetch,
    };
  }

  const icon: ThemeIconName = `weather.${weather.condition}`;

  return {
    weather,

    temperature: `${Math.round(weather.temperature)}ºC`,

    condition: weather.condition,

    icon,

    isDay: weather.isDay,

    isLoading,
    error,
    refetch,
  };
}
