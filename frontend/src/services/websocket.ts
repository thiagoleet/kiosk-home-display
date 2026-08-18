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

    const socket = this.socket;

    this.socket = null;

    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close();
    }
  }

  private createConnection(
    url: string,
    onMessage: WebSocketMessageHandler,
    onOpen?: WebSocketConnectionHandler,
    onClose?: WebSocketConnectionHandler,
  ) {
    if (this.manuallyDisconnected) {
      return;
    }

    const socket = new WebSocket(url);

    this.socket = socket;

    socket.onopen = () => {
      onOpen?.();
    };

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data) as WebSocketMessage;

      onMessage(message);
    };

    socket.onclose = () => {
      if (this.socket === socket) {
        this.socket = null;
      }

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
