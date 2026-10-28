export type NewsSourceStatusValue = "ok" | "error" | "pending";

export type NewsArticle = {
  id: string;
  title: string;
  summary?: string;
  source: string;
  category?: string;
  imageUrl?: string;
  url: string;
  publishedAt: string;
};

export type NewsSourceStatus = {
  name: string;
  category: string;
  status: NewsSourceStatusValue;
  lastSuccessAt: string | null;
};

// What GET /api/news returns. The daemon serialises empty Go slices as null,
// so the lists are not guaranteed to be arrays.
export type NewsFeedResponse = {
  articles: NewsArticle[] | null;
  updatedAt: string | null;
  sources: NewsSourceStatus[] | null;
};

// The feed once it has been checked. A device with news turned off answers
// 503, which is a setting rather than a failure, so it is a feed of its own.
export type NewsFeed = {
  isEnabled: boolean;
  articles: NewsArticle[];
  updatedAt: string | null;
  sources: NewsSourceStatus[];
};
