import { getWeatherForecast } from "@/services/weather-service";
import { useQuery } from "@tanstack/react-query";

export function useWeatherForecast() {
  return useQuery({
    queryKey: ["weather-forecast"],
    queryFn: getWeatherForecast,
  });
}
