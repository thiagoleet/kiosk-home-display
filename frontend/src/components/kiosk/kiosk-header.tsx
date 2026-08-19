import { Circle } from "lucide-react";

import { useKiosk } from "../../hooks/use-kiosk";
import type { WebSocketStatus } from "../../hooks/use-websocket";

type KioskHeaderProps = {
  status: WebSocketStatus;
};

export function KioskHeader({ status }: KioskHeaderProps) {
  const { profile } = useKiosk();

  const isConnected = status === "connected";

  return (
    <header className="kiosk-header">
      <h1 className="kiosk-name">{profile.name}</h1>

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
        />

        <span>{isConnected ? "ONLINE" : "OFFLINE"}</span>
      </div>
    </header>
  );
}
