import { useCallback, useEffect, useState } from "react";

type UseCarouselOptions = {
  length: number;
  interval?: number;
  isEnabled?: boolean;
};

const DEFAULT_INTERVAL = 15_000;

const DEFAULT_INDEX = 0;

export function useCarousel({
  length,
  interval = DEFAULT_INTERVAL,
  isEnabled = true,
}: UseCarouselOptions) {
  const [index, setActiveIndex] = useState(DEFAULT_INDEX);

  /**
   * Fall back to the default slide whenever the
   * active one is no longer available.
   */
  const activeIndex = index < length ? index : DEFAULT_INDEX;

  const goTo = useCallback(
    (index: number) => {
      if (length < 1) {
        return;
      }

      setActiveIndex(((index % length) + length) % length);
    },
    [length],
  );

  const next = useCallback(() => {
    goTo(activeIndex + 1);
  }, [goTo, activeIndex]);

  const previous = useCallback(() => {
    goTo(activeIndex - 1);
  }, [goTo, activeIndex]);

  /**
   * Rotate through the slides. A single slide never
   * rotates, so the default one simply stays put.
   */
  useEffect(() => {
    if (!isEnabled || length < 2 || interval <= 0) {
      return;
    }

    const timer = window.setInterval(() => {
      setActiveIndex((current) => (current + 1) % length);
    }, interval);

    return () => {
      window.clearInterval(timer);
    };
  }, [isEnabled, length, interval]);

  return {
    activeIndex,
    goTo,
    next,
    previous,
  };
}
