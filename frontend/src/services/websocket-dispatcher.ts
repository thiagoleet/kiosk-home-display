import type { WebSocketMessage } from "../types/websocket";

export type WebSocketMessageHandler = (message: WebSocketMessage) => void;

export class WebSocketDispatcher {
  private handlers = new Map<string, Set<WebSocketMessageHandler>>();

  subscribe(type: WebSocketMessage["type"], handler: WebSocketMessageHandler) {
    const handlers = this.handlers.get(type);

    if (handlers) {
      handlers.add(handler);
    } else {
      this.handlers.set(type, new Set([handler]));
    }

    return () => {
      this.unsubscribe(type, handler);
    };
  }

  unsubscribe(
    type: WebSocketMessage["type"],
    handler: WebSocketMessageHandler,
  ) {
    const handlers = this.handlers.get(type);

    if (!handlers) {
      return;
    }

    handlers.delete(handler);

    if (handlers.size === 0) {
      this.handlers.delete(type);
    }
  }

  dispatch(message: WebSocketMessage) {
    console.log("[WebSocket Dispatcher]", message);

    const handlers = this.handlers.get(message.type);

    if (!handlers) {
      console.log("[WebSocket Dispatcher] no handlers for", message.type);

      return;
    }

    handlers.forEach((handler) => {
      handler(message);
    });
  }
}
