import { useMemo } from "react";

import { Carousel } from "@/components/carousel/carousel";
import { HomeView } from "@/components/views/home-view";
import { WeatherView } from "@/components/views/weather-view";
import { ActivityView } from "../views/activity-view";

import { useCarousel } from "@/hooks/use-carousel";

import type { CarouselSlide } from "@/types/carousel";

function HomeEffects() {
  const items = Array.from({ length: 100 }, (_, index) => index);

  return (
    <>
      {items.map((item) => (
        <div
          key={item}
          className="circle-container"
        >
          <div className="circle"></div>
        </div>
      ))}
    </>
  );
}

export function HomeLayout() {
  /**
   * The home view is the first and default slide.
   * Extra slides are appended after it.
   */
  const slides = useMemo<CarouselSlide[]>(
    () => [
      { id: "home", content: <HomeView /> },
      { id: "weather", content: <WeatherView /> },
      { id: "activity", content: <ActivityView /> },
    ],
    [],
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

      <HomeEffects />
    </section>
  );
}
