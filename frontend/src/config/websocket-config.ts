// In production the URL derives from the page's own origin so it works behind the
// nginx reverse proxy in any deployment. During `vite dev` the frontend is served
// from Vite's port, so point straight at the backend instead.
// Set VITE_WS_URL to override either default.
const protocol = window.location.protocol === "https:" ? "wss" : "ws";

const defaultUrl = import.meta.env.DEV
  ? "ws://localhost:8080/ws"
  : `${protocol}://${window.location.host}/ws`;

export const websocketUrl = import.meta.env.VITE_WS_URL ?? defaultUrl;
