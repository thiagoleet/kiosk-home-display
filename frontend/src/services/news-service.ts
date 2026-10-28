import type {
  NewsArticle,
  NewsFeed,
  NewsSourceStatus,
  NewsSourceStatusValue,
} from "../types/news";

const NEWS_API_URL = "/api/news";

const DISABLED_FEED: NewsFeed = {
  isEnabled: false,
  articles: [],
  updatedAt: null,
  sources: [],
};

const SOURCE_STATUSES: NewsSourceStatusValue[] = ["ok", "error", "pending"];

export async function getNews(limit: number): Promise<NewsFeed> {
  const response = await fetch(`${NEWS_API_URL}?limit=${limit}`);

  // The daemon answers 503 when news is turned off for this device.
  if (response.status === 503) {
    return DISABLED_FEED;
  }

  if (!response.ok) {
    throw new Error(`Failed to fetch news: ${response.status}`);
  }

  return parseFeed(await response.json());
}

// The feed content comes from third-party RSS, so nothing in it is trusted to
// match the type: entries that cannot be shown are dropped instead of
// reaching the slideshow half-formed.
function parseFeed(data: unknown): NewsFeed {
  if (!isRecord(data)) {
    throw new Error("Unexpected news response");
  }

  return {
    isEnabled: true,
    articles: parseArticles(data.articles),
    updatedAt: parseTimestamp(data.updatedAt),
    sources: parseSources(data.sources),
  };
}

function parseArticles(value: unknown): NewsArticle[] {
  if (!Array.isArray(value)) {
    return [];
  }

  const seen = new Set<string>();
  const articles: NewsArticle[] = [];

  for (const item of value) {
    const article = parseArticle(item);

    if (!article || seen.has(article.id)) {
      continue;
    }

    seen.add(article.id);
    articles.push(article);
  }

  return articles;
}

function parseArticle(item: unknown): NewsArticle | null {
  if (!isRecord(item)) {
    return null;
  }

  const id = text(item.id);
  const title = text(item.title);

  if (!id || !title) {
    return null;
  }

  return {
    id,
    title,
    summary: text(item.summary) || undefined,
    source: text(item.source),
    category: text(item.category) || undefined,
    imageUrl: httpURL(item.imageUrl),
    url: httpURL(item.url) ?? "",
    publishedAt: parseTimestamp(item.publishedAt) ?? "",
  };
}

function parseSources(value: unknown): NewsSourceStatus[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.filter(isRecord).map((item) => ({
    name: text(item.name),
    category: text(item.category),
    status: SOURCE_STATUSES.includes(item.status as NewsSourceStatusValue)
      ? (item.status as NewsSourceStatusValue)
      : "pending",
    lastSuccessAt: parseTimestamp(item.lastSuccessAt),
  }));
}

function parseTimestamp(value: unknown): string | null {
  if (typeof value !== "string" || Number.isNaN(Date.parse(value))) {
    return null;
  }

  return value;
}

function httpURL(value: unknown): string | undefined {
  if (typeof value !== "string") {
    return undefined;
  }

  try {
    const url = new URL(value);

    return url.protocol === "http:" || url.protocol === "https:"
      ? url.href
      : undefined;
  } catch {
    return undefined;
  }
}

function text(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
