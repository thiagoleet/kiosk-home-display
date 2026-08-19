import type { Notification } from "../../types/notification";
import type { WebSocketStatus } from "../../types/websocket";
import { KioskClock } from "../kiosk/kiosk-clock";
import { KioskGreeting } from "../kiosk/kiosk-greeting";
import { KioskHeader } from "../kiosk/kiosk-header";
import { NotificationList } from "../notifications/notification-list";

type NotificationLayoutProps = {
  kioskName: string;
  connectionStatus: WebSocketStatus;
  notifications: Notification[];
};

export function NotificationLayout({
  kioskName,
  connectionStatus,
  notifications,
}: NotificationLayoutProps) {
  return (
    <section className="notification-layout">
      <KioskHeader
        name={kioskName}
        status={connectionStatus}
      />

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
