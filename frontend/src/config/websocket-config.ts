// Derives from the page's own origin so it works behind the nginx reverse proxy in any deployment.
// const protocol = window.location.protocol === "https:" ? "wss" : "ws";

// export const websocketUrl = `${protocol}://${window.location.host}/ws`;

export const websocketUrl = "ws://localhost:8080/ws";
