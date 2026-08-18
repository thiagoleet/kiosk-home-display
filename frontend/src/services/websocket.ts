import type { WebSocketMessage } from "../types/websocket";

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

export class WebSocketService {
  private socket: WebSocket | null = null;

  connect(
    url: string,
    onMessage: WebSocketMessageHandler,
    onOpen?: () => void,
    onClose?: () => void,
  ) {
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      onOpen?.();
    };

    this.socket.onmessage = (event) => {
      const message = JSON.parse(event.data) as WebSocketMessage;

      onMessage(message);
    };

    this.socket.onclose = () => {
      onClose?.();
    };
  }

  disconnect() {
    this.socket?.close();
    this.socket = null;
  }
}
