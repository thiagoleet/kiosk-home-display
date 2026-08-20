import { useCallback, useEffect, useState } from "react";

import type { Weather } from "../types/weather";
import { getWeather } from "../services/weather-service";

export function useWeather() {
  const [weather, setWeather] = useState<Weather | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchWeather = useCallback(async () => {
    try {
      setWeather(await getWeather());
      setError(null);
    } catch (error) {
      setError(
        error instanceof Error ? error : new Error("Failed to fetch weather"),
      );
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void fetchWeather();
    }, 0);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [fetchWeather]);

  return {
    weather,
    isLoading,
    error,
    refetch: fetchWeather,
  };
}
