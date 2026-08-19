import { createContext } from "react";

import type { KioskProfile } from "../types/kiosk-profile";

export type KioskContextValue = {
  profile: KioskProfile;
};

export const KioskContext = createContext<KioskContextValue | null>(null);
