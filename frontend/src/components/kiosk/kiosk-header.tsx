import { Bell, Circle } from "lucide-react";

import type { WebSocketStatus } from "../../hooks/use-websocket";
import { useKiosk } from "../../hooks/use-kiosk";

type KioskHeaderProps = {
  status: WebSocketStatus;
  hasNotification: boolean;
};

export function KioskHeader({ status, hasNotification }: KioskHeaderProps) {
  const isConnected = status === "connected";
  const { profile } = useKiosk();

  return (
    <header className="kiosk-header">
      <h1 className="kiosk-name">{profile.name}</h1>

      <div className="kiosk-header__status">
        {hasNotification && (
          <span
            className="notification-indicator"
            aria-label="Notification active"
          >
            <Bell
              size={18}
              aria-hidden="true"
            />

            <span
              className="notification-indicator__dot"
              aria-hidden="true"
            />
          </span>
        )}

        <div
          className={`connection-status ${
            isConnected
              ? "connection-status--connected"
              : "connection-status--disconnected"
          }`}
        >
          <Circle
            size={8}
            fill="currentColor"
            aria-hidden="true"
          />

          <span>{isConnected ? "ONLINE" : "OFFLINE"}</span>
        </div>
      </div>
    </header>
  );
}
