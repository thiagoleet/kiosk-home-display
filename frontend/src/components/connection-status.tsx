import type { WebSocketStatus } from "../hooks/use-websocket";

type ConnectionStatusProps = {
  status: WebSocketStatus;
};

export function ConnectionStatus({ status }: ConnectionStatusProps) {
  return <div>WebSocket: {status}</div>;
}
