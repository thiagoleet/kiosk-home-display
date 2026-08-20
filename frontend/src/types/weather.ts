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
  weatherCode: number;
  isDay: boolean;
  timestamp: string;
};
