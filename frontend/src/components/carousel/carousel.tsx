import type { CSSProperties } from "react";

import type { CarouselSlide } from "@/types/carousel";

type CarouselProps = {
  slides: CarouselSlide[];
  activeIndex: number;
  hasIndicators?: boolean;
};

type CarouselTrackStyle = CSSProperties & {
  "--carousel-index": number;
};

export function Carousel({
  slides,
  activeIndex,
  hasIndicators = true,
}: CarouselProps) {
  if (slides.length === 0) {
    return null;
  }

  const trackStyle: CarouselTrackStyle = {
    "--carousel-index": activeIndex,
  };

  return (
    <div className="carousel">
      <div
        className="carousel__track"
        style={trackStyle}
      >
        {slides.map((slide, index) => (
          <div
            key={slide.id}
            className="carousel__slide"
            data-active={index === activeIndex}
            aria-hidden={index !== activeIndex}
          >
            {slide.content}
          </div>
        ))}
      </div>

      {hasIndicators && slides.length > 1 && (
        <div className="carousel__indicators">
          {slides.map((slide, index) => (
            <span
              key={slide.id}
              className="carousel__indicator"
              data-active={index === activeIndex}
            />
          ))}
        </div>
      )}
    </div>
  );
}
