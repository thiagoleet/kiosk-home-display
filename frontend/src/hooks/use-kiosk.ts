import { useContext } from "react";

import { KioskContext } from "../providers/kiosk-context";

export function useKiosk() {
  const context = useContext(KioskContext);

  if (!context) {
    throw new Error("useKiosk must be used inside KioskProvider");
  }

  return context;
}
