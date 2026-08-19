// Components
import { KioskLayout } from "./components/kiosk/kiosk-layout";
import { KioskScreen } from "./components/kiosk/kiosk-screen";

// Hooks
import { useKioskState } from "./hooks/use-kiosk-state";
import { useNotifications } from "./hooks/use-notifications";
import { usePrinter } from "./hooks/use-printer";
import { useWebSocketContext } from "./hooks/use-websocket-context";
import { useScreenMode } from "./hooks/use-screen-mode";

import { kioskProfile } from "./profiles";

function App() {
  const { state } = useKioskState();
  const { notifications } = useNotifications();
  const { state: printerState, currentJob } = usePrinter();
  const { status } = useWebSocketContext();

  const { mode } = useScreenMode({
    notifications,
  });

  return (
    <KioskScreen mode={mode}>
      <KioskLayout
        mode={mode}
        kioskName={kioskProfile.name}
        connectionStatus={status}
        printerState={printerState}
        currentJob={currentJob}
        notifications={notifications}
        activities={[]}
      />
    </KioskScreen>
  );
}

export default App;
