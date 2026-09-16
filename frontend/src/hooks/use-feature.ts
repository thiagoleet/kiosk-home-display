import { useKiosk } from "./use-kiosk";

import type { KioskProfile } from "../types/kiosk-profile";

type Feature = keyof KioskProfile["features"];

// Whether the box this build targets has the feature at all. A feature the
// profile turns off has no hardware behind it, so the surfaces that speak for
// it are left out rather than shown idle.
export function useFeature(feature: Feature) {
  const { profile } = useKiosk();

  return profile.features[feature];
}
