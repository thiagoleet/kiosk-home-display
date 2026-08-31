import { useEffect, useState } from "react";

import { useKiosk } from "@/hooks/use-kiosk";
import { useTranslation } from "@/hooks/use-translation";
import { useWebSocketContext } from "@/hooks/use-websocket-context";
import { ThemeIcon } from "../theme/theme-icon";

const ONLINE_STATUS_DURATION = 3000;

type KioskHeaderProps = {
  hasNotification: boolean;
};

type ConnectionStatusIndicatorProps = {
  isConnected: boolean;
};

type NotificationIndicatorProps = {
  hasNotification: boolean;
};

const ConnectionStatusIndicator = ({
  isConnected,
}: ConnectionStatusIndicatorProps) => {
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
      <ThemeIcon
        name={isConnected ? "status.online" : "status.offline"}
        size={8}
        fill="currentColor"
        aria-hidden="true"
      />

      <span>{t(isConnected ? "status.online" : "status.offline")}</span>
    </div>
  );
};

const NotificationIndicator = ({
  hasNotification,
}: NotificationIndicatorProps) => {
  const { t } = useTranslation();

  if (!hasNotification) {
    return null;
  }

  return (
    <span
      className="notification-indicator"
      aria-label={t("notification.active")}
    >
      <ThemeIcon
        name="notification"
        aria-hidden="true"
      />

      <span
        className="notification-indicator__dot"
        aria-hidden="true"
      />
    </span>
  );
};

export function KioskHeader({ hasNotification }: KioskHeaderProps) {
  const { status } = useWebSocketContext();

  const isConnected = status === "connected";
  const { profile } = useKiosk();

  return (
    <header className="kiosk-header">
      <h1 className="kiosk-name">{profile.name}</h1>

      <div className="kiosk-header__status">
        <NotificationIndicator hasNotification={hasNotification} />
        <ConnectionStatusIndicator isConnected={isConnected} />
      </div>
    </header>
  );
}
