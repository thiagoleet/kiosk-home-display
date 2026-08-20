// Components
import { KioskLayout } from "./components/kiosk/kiosk-layout";
import { KioskScreen } from "./components/kiosk/kiosk-screen";

// Hooks

import { useNotifications } from "./hooks/use-notifications";
import { useScreenMode } from "./hooks/use-screen-mode";
import { useKioskState } from "./hooks/use-kiosk-state";

function App() {
  const { activeNotification } = useNotifications();
  const { state } = useKioskState();

  const { mode } = useScreenMode({
    notification: activeNotification,
    isScreenOn: state.display.power === "on",
  });

  return (
    <KioskScreen mode={mode}>
      <KioskLayout
        mode={mode}
        notification={activeNotification}
      />
    </KioskScreen>
  );
}

export default App;
