import { Bell, Circle } from "lucide-react";
import { useEffect, useState } from "react";

import { useKiosk } from "../../hooks/use-kiosk";
import { useTranslation } from "../../hooks/use-translation";
import { useWebSocketContext } from "../../hooks/use-websocket-context";

type KioskHeaderProps = {
  hasNotification: boolean;
};

const ONLINE_STATUS_DURATION = 3000;

export function KioskHeader({ hasNotification }: KioskHeaderProps) {
  const { status } = useWebSocketContext();

  const isConnected = status === "connected";
  const { profile } = useKiosk();
  const { t } = useTranslation();
  const [isConnectionStatusVisible, setConnectionStatusVisible] =
    useState(true);
  const shouldShowConnectionStatus = !isConnected || isConnectionStatusVisible;

  useEffect(() => {
    if (!isConnected) {
      return;
    }

    const showTimer = window.setTimeout(() => {
      setConnectionStatusVisible(true);
    }, 0);

    const hideTimer = window.setTimeout(() => {
      setConnectionStatusVisible(false);
    }, ONLINE_STATUS_DURATION);

    return () => {
      window.clearTimeout(showTimer);
      window.clearTimeout(hideTimer);
    };
  }, [isConnected]);

  return (
    <header className="kiosk-header">
      <h1 className="kiosk-name">{profile.name}</h1>

      <div className="kiosk-header__status">
        {hasNotification && (
          <span
            className="notification-indicator"
            aria-label={t("notification.active")}
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
          className={[
            "connection-status",
            isConnected
              ? "connection-status--connected"
              : "connection-status--disconnected",
            !shouldShowConnectionStatus && "connection-status--hidden",
          ]
            .filter(Boolean)
            .join(" ")}
          aria-hidden={!shouldShowConnectionStatus}
        >
          <Circle
            size={8}
            fill="currentColor"
            aria-hidden="true"
          />

          <span>{t(isConnected ? "status.online" : "status.offline")}</span>
        </div>
      </div>
    </header>
  );
}
