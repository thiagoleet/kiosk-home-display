import type { WebSocketMessage } from "../types/websocket";

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

export type WebSocketConnectionHandler = () => void;

type WebSocketServiceOptions = {
  reconnectDelay?: number;
};

export class WebSocketService {
  private socket: WebSocket | null = null;
  private reconnectTimer: number | null = null;

  private readonly reconnectDelay: number;

  private manuallyDisconnected = false;

  constructor(options: WebSocketServiceOptions = {}) {
    this.reconnectDelay = options.reconnectDelay ?? 2000;
  }

  connect(
    url: string,
    onMessage: WebSocketMessageHandler,
    onOpen?: WebSocketConnectionHandler,
    onClose?: WebSocketConnectionHandler,
  ) {
    this.manuallyDisconnected = false;

    this.createConnection(url, onMessage, onOpen, onClose);
  }

  disconnect() {
    this.manuallyDisconnected = true;

    this.clearReconnectTimer();

    this.socket?.close();
    this.socket = null;
  }

  private createConnection(
    url: string,
    onMessage: WebSocketMessageHandler,
    onOpen?: WebSocketConnectionHandler,
    onClose?: WebSocketConnectionHandler,
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
      this.socket = null;

      onClose?.();

      if (!this.manuallyDisconnected) {
        this.scheduleReconnect(url, onMessage, onOpen, onClose);
      }
    };
  }

  private scheduleReconnect(
    url: string,
    onMessage: WebSocketMessageHandler,
    onOpen?: WebSocketConnectionHandler,
    onClose?: WebSocketConnectionHandler,
  ) {
    this.clearReconnectTimer();

    this.reconnectTimer = window.setTimeout(() => {
      this.createConnection(url, onMessage, onOpen, onClose);
    }, this.reconnectDelay);
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer);

      this.reconnectTimer = null;
    }
  }
}
