import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./styles/index.css";
import App from "./app";
import { WebSocketProvider } from "./providers/websocket-provider";
import { KioskProvider } from "./providers/kiosk-provider";
import { ThemeProvider } from "./providers/theme-provider";
import { I18nProvider } from "./providers/i18n-provider";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <KioskProvider>
      <I18nProvider>
        <ThemeProvider>
          <WebSocketProvider>
            <App />
          </WebSocketProvider>
        </ThemeProvider>
      </I18nProvider>
    </KioskProvider>
  </StrictMode>,
);
