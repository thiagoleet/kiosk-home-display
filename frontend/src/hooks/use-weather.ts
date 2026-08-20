import { useQuery } from "@tanstack/react-query";
import { getWeather } from "../services/weather-service";

export function useWeather() {
  return useQuery({
    queryKey: ["weather"],
    queryFn: getWeather,
  });
}
