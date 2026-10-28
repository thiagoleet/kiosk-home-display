// Every article is a slide of its own, sharing the carousel's interval with the
// other home slides, so the count is kept low enough that news does not take
// over the rotation.
export const NEWS_LIMIT = 6;

// The daemon refreshes its feeds every 20 minutes by default; polling at half
// that picks up a refresh without asking for the same snapshot over and over.
export const NEWS_REFRESH_INTERVAL = 10 * 60_000;
