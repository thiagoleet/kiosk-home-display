import type { KioskProfile } from "../types/kiosk-profile";

// SNESPI runs the same screen as MILKPI without a printer attached to it. The
// daemon on that box runs with PRINTER_MODE=off, so no print event ever
// reaches the frontend; this flag keeps the printer surfaces out of the UI so
// the screen never offers something the host cannot do.
export const snespiProfile: KioskProfile = {
  id: "snespi",
  name: "SNESPI",

  features: {
    printer: false,
  },

  theme: "retro-16bit",
  locale: "pt-BR",
};
