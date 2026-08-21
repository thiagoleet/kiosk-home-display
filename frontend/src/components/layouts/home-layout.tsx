import { useMemo } from "react";

import { Carousel } from "@/components/carousel/carousel";
import { HomeView } from "@/components/views/home-view";

import { ActivityWidget } from "@/components/widgets/activity-widget";
import { PrinterWidget } from "@/components/widgets/printer-widget";

import { useCarousel } from "@/hooks/use-carousel";

import type { CarouselSlide } from "@/types/carousel";

export function HomeLayout() {
  /**
   * The home view is the first and default slide.
   * Extra slides are appended after it.
   */
  const slides = useMemo<CarouselSlide[]>(
    () => [{ id: "home", content: <HomeView /> }],
    [],
  );

  const { activeIndex } = useCarousel({ length: slides.length });

  return (
    <section className="home-layout">
      <div className="home-layout__main">
        <Carousel
          slides={slides}
          activeIndex={activeIndex}
        />
      </div>

      <footer className="home-layout__footer">
        <PrinterWidget />
        <ActivityWidget />
      </footer>
    </section>
  );
}
