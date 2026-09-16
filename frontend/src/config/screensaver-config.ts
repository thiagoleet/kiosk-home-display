// The screensaver is on by default. Set VITE_SCREENSAVER=off to keep the kiosk
// on the home screen no matter what the daemon reports about idling or display
// power — useful while developing, or on a panel that should never blank.
// Vite inlines this at build time, so changing it means rebuilding.
export const screensaverEnabled = import.meta.env.VITE_SCREENSAVER !== "off";
