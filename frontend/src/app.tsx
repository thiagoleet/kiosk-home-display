import { ConnectionStatus } from "./components/connection-status";
import { DisplayStatus } from "./components/display-status";
import { NotificationList } from "./components/notification-list";
import { PrinterStatus } from "./components/printer-status";
import { useKioskState } from "./hooks/use-kiosk-state";
import { useNotifications } from "./hooks/use-notifications";
import { usePrinter } from "./hooks/use-printer";
import { useWebSocketContext } from "./hooks/use-websocket-context";

function App() {
  const { state } = useKioskState();

  const { notifications } = useNotifications();

  const { status } = useWebSocketContext();

  const { state: printerState, currentJob } = usePrinter();

  return (
    <main>
      <h1>Kiosk Home Display</h1>

      <ConnectionStatus status={status} />

      <DisplayStatus display={state.display} />

      <PrinterStatus
        state={printerState}
        job={currentJob}
      />

      <NotificationList notifications={notifications} />
    </main>
  );
}

export default App;
