import type { ThemeIconName } from "./theme";

export type WeatherCondition =
  | "clear"
  | "partly_cloudy"
  | "overcast"
  | "fog"
  | "drizzle"
  | "rain"
  | "rain_showers"
  | "snow"
  | "thunderstorm";

export type Weather = {
  temperature: number;
  apparentTemperature: number;
  humidity: number;
  windSpeed: number;
  condition: WeatherCondition;
  isDay: boolean;
  timestamp: string;
};

export type WeatherForecastResponse = {
  date: string;
  temperatureMin: number;
  temperatureMax: number;
  precipitationProbability: number;
  condition: WeatherCondition;
};

export type WeatherForecast = {
  icon: ThemeIconName;
  temperatureMin: string;
  temperatureMax: string;
} & Omit<WeatherForecastResponse, "temperatureMin" | "temperatureMax">;
