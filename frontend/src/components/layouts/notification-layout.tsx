import type { Notification } from "../../types/notification";
import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { NotificationList } from "../notifications/notification-list";

type NotificationLayoutProps = {
  notification: Notification | null;
};

export function NotificationLayout({ notification }: NotificationLayoutProps) {
  return (
    <section className="notification-layout">
      <div className="notification-layout__content">
        <div className="notification-layout__info">
          <KioskGreeting />

          <KioskClock />
        </div>

        <div className="notification-layout__panel">
          <NotificationList notification={notification} />
        </div>
      </div>
    </section>
  );
}
