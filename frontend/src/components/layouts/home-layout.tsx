import { useMemo } from "react";

import { Carousel } from "@/components/carousel/carousel";
import { Particles } from "../effects/particles";
import { HomeView } from "@/components/views/home-view";
import { WeatherView } from "@/components/views/weather-view";
import { ActivityView } from "../views/activity-view";
import { WeatherForecastViewWidget } from "../view-widgets/weather-forecast-view-widget";
import { NewsView } from "../views/news-view";

import { useActivities } from "@/hooks/use-activities";
import { useCarousel } from "@/hooks/use-carousel";
import { useNews } from "@/hooks/use-news";

import type { CarouselSlide } from "@/types/carousel";

const MAX_ACTIVITY_SLIDE_ITEMS = 8;

export function HomeLayout() {
  /**
   * Activities are loaded here rather than inside the
   * slide, so the subscription keeps running while the
   * slide is hidden and brings it back on the first event.
   */
  const { activities } = useActivities({
    maxActivities: MAX_ACTIVITY_SLIDE_ITEMS,
  });

  const { articles } = useNews();

  const hasActivities = activities.length > 0;

  /**
   * The home view is the first and default slide.
   * Extra slides are appended after it. Activity only
   * shows when there is something to list, and news goes
   * last, so the slides before them keep their positions
   * when either one changes size or goes away.
   */
  const slides = useMemo<CarouselSlide[]>(
    () => [
      { id: "home", content: <HomeView /> },
      { id: "weather", content: <WeatherView /> },
      { id: "weather-forecast", content: <WeatherForecastViewWidget /> },
      ...(hasActivities
        ? [
            {
              id: "activity",
              content: <ActivityView activities={activities} />,
            },
          ]
        : []),
      ...articles.map((article) => ({
        id: `news-${article.id}`,
        content: <NewsView article={article} />,
      })),
    ],
    [hasActivities, activities, articles],
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
