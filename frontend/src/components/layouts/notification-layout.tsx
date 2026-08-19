import type { Notification } from "../../types/notification";
import type { WebSocketStatus } from "../../types/websocket";
import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskHeader } from "../kiosk/kiosk-header";
import { NotificationList } from "../notifications/notification-list";

type NotificationLayoutProps = {
  connectionStatus: WebSocketStatus;
  notifications: Notification[];
};

export function NotificationLayout({
  connectionStatus,
  notifications,
}: NotificationLayoutProps) {
  return (
    <section className="notification-layout">
      <KioskHeader status={connectionStatus} />

      <div className="notification-layout__content">
        <div className="notification-layout__info">
          <KioskGreeting />

          <KioskClock />
        </div>

        <div className="notification-layout__panel">
          <NotificationList notifications={notifications} />
        </div>
      </div>
    </section>
  );
}
