import { KioskHeader } from "./kiosk-header";
import { KioskTransition } from "./kiosk-transition";

import { HomeLayout } from "../layouts/home-layout";
import { NotificationLayout } from "../layouts/notification-layout";
import { ScreenSaver } from "../screensaver/screensaver";

import type { Notification } from "../../types/notification";
import type { ScreenMode } from "../../types/screen";
import { Activity } from "react";

type KioskLayoutProps = {
  mode: ScreenMode;
  notification: Notification | null;
};

export function KioskLayout({ mode, notification }: KioskLayoutProps) {
  return (
    <section className="kiosk-layout">
      <Activity mode={mode === "screensaver" ? "hidden" : "visible"}>
        <KioskHeader hasNotification={notification !== null} />
      </Activity>

      <div className="kiosk-layout__content">
        <KioskTransition
          mode={mode}
          transitionKey={
            mode === "notification"
              ? (notification?.id ?? "notification")
              : mode === "screensaver"
                ? "screensaver"
                : "home"
          }
        >
          {mode === "screensaver" ? (
            <ScreenSaver />
          ) : mode === "notification" && notification ? (
            <NotificationLayout notification={notification} />
          ) : (
            <HomeLayout />
          )}
        </KioskTransition>
      </div>
    </section>
  );
}
