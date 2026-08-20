import { useWeather } from "./use-weather";

import type { WeatherCondition } from "../types/weather";

function getWeatherCondition(weatherCode: number): WeatherCondition {
  switch (weatherCode) {
    case 0:
      return "clear";

    case 1:
    case 2:
      return "partly_cloudy";

    case 3:
      return "overcast";

    case 45:
    case 48:
      return "fog";

    case 51:
    case 53:
    case 55:
    case 56:
    case 57:
      return "drizzle";

    case 61:
    case 63:
    case 65:
    case 66:
    case 67:
      return "rain";

    case 71:
    case 73:
    case 75:
    case 77:
    case 85:
    case 86:
      return "snow";

    case 80:
    case 81:
    case 82:
      return "rain_showers";

    case 95:
    case 96:
    case 99:
      return "thunderstorm";

    default:
      return "clear";
  }
}

function formatTemperature(temperature: number): string {
  return `${Math.round(temperature)}ºC`;
}

export function useWeatherWidget() {
  const { data: weather, isLoading, error, refetch } = useWeather();

  if (!weather) {
    return {
      weather: null,
      temperature: null,
      condition: null,
      isDay: null,
      icon: null,
      isLoading,
      error,
      refetch,
    };
  }

  const condition = getWeatherCondition(weather.weatherCode);

  return {
    weather,
    temperature: formatTemperature(weather.temperature),
    condition,
    isDay: weather.isDay,
    icon: `weather.${condition}`,
    isLoading,
    error,
    refetch,
  };
}
