import type { PropsWithChildren } from "react";

import { KioskContext } from "./kiosk-context";
import { kioskProfile } from "../profiles";

export function KioskProvider({ children }: PropsWithChildren) {
  return (
    <KioskContext.Provider
      value={{
        profile: kioskProfile,
      }}
    >
      {children}
    </KioskContext.Provider>
  );
}
