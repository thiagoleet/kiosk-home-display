import { ConnectionStatus } from "./components/connection-status";
import { DisplayStatus } from "./components/display-status";
import { NotificationList } from "./components/notification-list";
import { useKioskState } from "./hooks/use-kiosk-state";
import { useNotifications } from "./hooks/use-notifications";
import { useWebSocketContext } from "./hooks/use-websocket-context";

function App() {
  const { state } = useKioskState();

  const { notifications } = useNotifications();

  const { status } = useWebSocketContext();

  return (
    <main>
      <h1>Kiosk Home Display</h1>

      <ConnectionStatus status={status} />

      <DisplayStatus display={state.display} />

      <NotificationList notifications={notifications} />
    </main>
  );
}

export default App;
