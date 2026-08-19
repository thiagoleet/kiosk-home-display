import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./app";
import { WebSocketProvider } from "./providers/websocket-provider";
import { KioskProvider } from "./providers/kiosk-provider";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <KioskProvider>
      <WebSocketProvider>
        <App />
      </WebSocketProvider>
    </KioskProvider>
  </StrictMode>,
);
