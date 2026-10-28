import { useQuery } from "@tanstack/react-query";

import { getNews } from "../services/news-service";

import { NEWS_LIMIT, NEWS_REFRESH_INTERVAL } from "../constants/news";
import type { NewsArticle } from "../types/news";

const NO_ARTICLES: NewsArticle[] = [];

// A failed refresh keeps the last feed that loaded, so a blip on the daemon
// does not pull slides out from under the carousel; a 503 is a successful
// answer and clears them.
export function useNews(limit = NEWS_LIMIT) {
  const { data, isLoading, error } = useQuery({
    queryKey: ["news", limit],
    queryFn: () => getNews(limit),
    staleTime: NEWS_REFRESH_INTERVAL,
    refetchInterval: NEWS_REFRESH_INTERVAL,
  });

  return {
    articles: data?.isEnabled ? data.articles : NO_ARTICLES,
    updatedAt: data?.updatedAt ?? null,
    isEnabled: data?.isEnabled ?? false,
    isLoading,
    error,
  };
}
