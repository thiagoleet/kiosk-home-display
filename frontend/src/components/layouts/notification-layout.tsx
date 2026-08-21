import type { Notification } from "@/types/notification";
import { KioskViewWidget } from "../view-widgets/clock-view-widget";
import { GreetingViewWidget } from "../view-widgets/greeting-view-widget";
import { NotificationList } from "../notifications/notification-list";

type NotificationLayoutProps = {
  notification: Notification | null;
};

export function NotificationLayout({ notification }: NotificationLayoutProps) {
  return (
    <section className="notification-layout">
      <div className="notification-layout__content">
        <div className="notification-layout__info">
          <GreetingViewWidget />

          <KioskViewWidget />
        </div>

        <div className="notification-layout__panel">
          <NotificationList notification={notification} />
        </div>
      </div>
    </section>
  );
}
