import { Circle } from "lucide-react";

import type { WebSocketStatus } from "../../types/websocket";

type KioskHeaderProps = {
  name: string;
  status: WebSocketStatus;
};

export function KioskHeader({ name, status }: KioskHeaderProps) {
  const isConnected = status === "connected";

  return (
    <header className="kiosk-header">
      <h1 className="kiosk-name">{name}</h1>

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
