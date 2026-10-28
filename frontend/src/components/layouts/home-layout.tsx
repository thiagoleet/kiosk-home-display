import { useMemo } from "react";

import { Carousel } from "@/components/carousel/carousel";
import { Particles } from "../effects/particles";
import { HomeView } from "@/components/views/home-view";
import { WeatherView } from "@/components/views/weather-view";
import { ActivityView } from "../views/activity-view";
import { WeatherForecastViewWidget } from "../view-widgets/weather-forecast-view-widget";
import { NewsView } from "../views/news-view";

import { useCarousel } from "@/hooks/use-carousel";
import { useNews } from "@/hooks/use-news";

import type { CarouselSlide } from "@/types/carousel";

export function HomeLayout() {
  const { articles } = useNews();

  /**
   * The home view is the first and default slide.
   * Extra slides are appended after it. News goes last,
   * so the fixed slides keep their positions when the
   * feed changes size or goes away.
   */
  const slides = useMemo<CarouselSlide[]>(
    () => [
      { id: "home", content: <HomeView /> },
      { id: "weather", content: <WeatherView /> },
      { id: "weather-forecast", content: <WeatherForecastViewWidget /> },
      { id: "activity", content: <ActivityView /> },
      ...articles.map((article) => ({
        id: `news-${article.id}`,
        content: <NewsView article={article} />,
      })),
    ],
    [articles],
  );

  const { activeIndex, goTo } = useCarousel({ length: slides.length });

  return (
    <section className="home-layout">
      <div className="home-layout__main">
        <Carousel
          slides={slides}
          activeIndex={activeIndex}
          onIndicatorClick={goTo}
        />
      </div>

      <Particles />
    </section>
  );
}
