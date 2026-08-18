import { useEffect, useRef, useState } from "react";

import {
  WebSocketService,
  type WebSocketMessageHandler,
} from "../services/websocket";

export type WebSocketStatus = "connecting" | "connected" | "disconnected";

type UseWebSocketOptions = {
  url: string;
  onMessage: WebSocketMessageHandler;
};

export function useWebSocket({ url, onMessage }: UseWebSocketOptions) {
  const serviceRef = useRef<WebSocketService | null>(null);

  const [status, setStatus] = useState<WebSocketStatus>("connecting");

  useEffect(() => {
    console.log("WebSocket effect: connect");

    const service = new WebSocketService();
    serviceRef.current = service;

    service.connect(
      url,
      onMessage,
      () => {
        setStatus("connected");
      },
      () => {
        setStatus("disconnected");
      },
    );

    return () => {
      console.log("WebSocket effect: disconnect");

      service.disconnect();
      serviceRef.current = null;
    };
  }, [url, onMessage]);

  return {
    status,
  };
}
