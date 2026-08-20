import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import "./styles/index.css";
import App from "./app";
import { WebSocketProvider } from "./providers/websocket-provider";
import { KioskProvider } from "./providers/kiosk-provider";
import { ThemeProvider } from "./providers/theme-provider";
import { I18nProvider } from "./providers/i18n-provider";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <KioskProvider>
        <I18nProvider>
          <ThemeProvider>
            <WebSocketProvider>
              <App />
            </WebSocketProvider>
          </ThemeProvider>
        </I18nProvider>
      </KioskProvider>
    </QueryClientProvider>
  </StrictMode>,
);
