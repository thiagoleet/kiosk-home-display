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
    const service = new WebSocketService({
      reconnectDelay: 2000,
    });

    serviceRef.current = service;

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setStatus("connecting");

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
      service.disconnect();
      serviceRef.current = null;
    };
  }, [url, onMessage]);

  return {
    status,
  };
}
