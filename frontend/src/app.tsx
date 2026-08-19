// Components
import { ConnectionStatus } from "./components/connection-status";
import { DisplayStatus } from "./components/display-status";
import { NotificationList } from "./components/notifications/notification-list";
import { PrinterStatus } from "./components/printer-status";
import { KioskLayout } from "./components/kiosk-layout";
import { KioskScreen } from "./components/kiosk-screen";

// Hooks
import { useKioskState } from "./hooks/use-kiosk-state";
import { useNotifications } from "./hooks/use-notifications";
import { usePrinter } from "./hooks/use-printer";
import { useWebSocketContext } from "./hooks/use-websocket-context";
import { useScreenMode } from "./hooks/use-screen-mode";

function App() {
  const { state } = useKioskState();
  const { notifications } = useNotifications();
  const { status } = useWebSocketContext();
  const { state: printerState, currentJob } = usePrinter();
  const { mode } = useScreenMode({
    notifications,
  });

  return (
    <KioskScreen mode={mode}>
      <KioskLayout mode={mode}>
        <ConnectionStatus status={status} />

        <DisplayStatus display={state.display} />

        <PrinterStatus
          state={printerState}
          job={currentJob}
        />

        <NotificationList notifications={notifications} />
      </KioskLayout>
    </KioskScreen>
  );
}

export default App;
