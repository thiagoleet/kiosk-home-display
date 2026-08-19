// Components
import { KioskLayout } from "./components/kiosk/kiosk-layout";
import { KioskScreen } from "./components/kiosk/kiosk-screen";

// Hooks

import { useNotifications } from "./hooks/use-notifications";
import { usePrinter } from "./hooks/use-printer";
import { useWebSocketContext } from "./hooks/use-websocket-context";
import { useScreenMode } from "./hooks/use-screen-mode";
import { useActivities } from "./hooks/use-activities";

function App() {
  const { activeNotification } = useNotifications();
  const { activities } = useActivities();
  const { state: printerState, currentJob } = usePrinter();
  const { status } = useWebSocketContext();

  const { mode } = useScreenMode({
    notification: activeNotification,
  });

  return (
    <KioskScreen mode={mode}>
      <KioskLayout
        mode={mode}
        connectionStatus={status}
        printerState={printerState}
        currentJob={currentJob}
        notification={activeNotification}
        activities={activities}
      />
    </KioskScreen>
  );
}

export default App;
