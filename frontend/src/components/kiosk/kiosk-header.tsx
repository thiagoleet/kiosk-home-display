import { useEffect, useMemo, useState } from "react";

// import { useKiosk } from "@/hooks/use-kiosk";
import { useTranslation } from "@/hooks/use-translation";
import { useWebSocketContext } from "@/hooks/use-websocket-context";
import { ThemeIcon } from "../theme/theme-icon";
import type { WebSocketStatus } from "@/types/websocket";

const ONLINE_STATUS_DURATION = 3000;

type KioskHeaderProps = {
  hasNotification: boolean;
};

type ConnectionStatusIndicatorProps = {
  status: WebSocketStatus;
};

type NotificationIndicatorProps = {
  hasNotification: boolean;
};

type ConnectionStatus = {
  className: string;
  iconName: "status.online" | "status.offline" | "status.connecting";
};

const ConnectionStatusIndicator = ({
  status,
}: ConnectionStatusIndicatorProps) => {
  const { t } = useTranslation();

  const [isConnectionStatusVisible, setConnectionStatusVisible] =
    useState(true);
  const shouldShowConnectionStatus =
    status === "connected" || isConnectionStatusVisible;

  const statusClass = useMemo<ConnectionStatus>(() => {
    switch (status) {
      case "connected":
        return {
          className: "connection-status--connected",
          iconName: "status.online",
        };
      case "disconnected":
        return {
          className: "connection-status--disconnected",
          iconName: "status.offline",
        };
      case "connecting":
        return {
          className: "connection-status--connecting",
          iconName: "status.connecting",
        };
      default:
        return {
          className: "connection-status--disconnected",
          iconName: "status.offline",
        };
    }
  }, [status]);

  useEffect(() => {
    if (status === "disconnected") {
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
  }, [status]);

  return (
    <div
      className={[
        "connection-status",
        statusClass.className,
        !shouldShowConnectionStatus && "connection-status--hidden",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-hidden={!shouldShowConnectionStatus}
    >
      <ThemeIcon
        name={statusClass.iconName}
        size={16}
        fill="currentColor"
        aria-hidden="true"
      />

      <span>{t(statusClass.iconName)}</span>
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

  return (
    <header className="kiosk-header">
      {/* <h1 className="kiosk-name">{profile.name}</h1> */}
      <div className="kiosk-name"></div>

      <div className="kiosk-header__status">
        <NotificationIndicator hasNotification={hasNotification} />
        <ConnectionStatusIndicator status={status} />
      </div>
    </header>
  );
}
