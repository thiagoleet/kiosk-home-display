// Which box this build is for. Set VITE_KIOSK_PROFILE at build time — see
// deploy/environments/<name>/frontend.env for the value each Pi is built with.
// Vite inlines it, so every box needs its own build.
export const kioskProfileId = import.meta.env.VITE_KIOSK_PROFILE ?? "";
