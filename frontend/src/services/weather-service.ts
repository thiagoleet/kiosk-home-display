import type { Weather } from "../types/weather";

const WEATHER_API_URL = "/api/weather";

export async function getWeather(): Promise<Weather> {
  const response = await fetch(WEATHER_API_URL);

  if (!response.ok) {
    throw new Error(`Failed to fetch weather: ${response.status}`);
  }

  return response.json();
}
