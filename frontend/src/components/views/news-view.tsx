import { useState } from "react";

import newsFallbackImage from "@/assets/images/news-fallback.jpeg";

import type { NewsArticle } from "@/types/news";

type NewsViewProps = Readonly<{
  article: NewsArticle;
}>;

type NewsImageProps = Readonly<{
  src?: string;
}>;

// Falls back to the bundled image when the article has none or the browser
// cannot load it. The fallback itself is never retried, so a missing asset
// cannot loop.
function NewsImage({ src }: NewsImageProps) {
  const [hasFailed, setHasFailed] = useState(false);

  const imageSrc = src && !hasFailed ? src : newsFallbackImage;

  return (
    <img
      className="news-view__image"
      src={imageSrc}
      alt=""
      decoding="async"
      draggable={false}
      onError={() => {
        if (imageSrc !== newsFallbackImage) {
          setHasFailed(true);
        }
      }}
    />
  );
}

export function NewsView({ article }: NewsViewProps) {
  return (
    <article className="news-view">
      <NewsImage
        key={article.imageUrl ?? "fallback"}
        src={article.imageUrl}
      />

      <div className="news-view__caption">
        <h2 className="news-view__headline">{article.title}</h2>

        {article.source && (
          <span className="news-view__source">{article.source}</span>
        )}
      </div>
    </article>
  );
}
