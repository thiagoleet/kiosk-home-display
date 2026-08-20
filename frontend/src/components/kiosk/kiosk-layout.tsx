import { KioskHeader } from "./kiosk-header";
import { KioskTransition } from "./kiosk-transition";

import { HomeLayout } from "../layouts/home-layout";
import { NotificationLayout } from "../layouts/notification-layout";

import type { Notification } from "../../types/notification";
import type { ScreenMode } from "../../types/screen";

type KioskLayoutProps = {
  mode: ScreenMode;
  notification: Notification | null;
};

export function KioskLayout({ mode, notification }: KioskLayoutProps) {
  return (
    <section className="kiosk-layout">
      <KioskHeader hasNotification={notification !== null} />

      <div className="kiosk-layout__content">
        <KioskTransition
          mode={mode}
          transitionKey={
            mode === "notification"
              ? (notification?.id ?? "notification")
              : "home"
          }
        >
          {mode === "notification" && notification ? (
            <NotificationLayout notification={notification} />
          ) : (
            <HomeLayout />
          )}
        </KioskTransition>
      </div>
    </section>
  );
}
