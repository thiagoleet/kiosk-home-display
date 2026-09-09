// Components
import { KioskLayout } from "./components/kiosk/kiosk-layout";
import { KioskScreen } from "./components/kiosk/kiosk-screen";

// Hooks

import { useNotifications } from "./hooks/use-notifications";
import { useScreenMode } from "./hooks/use-screen-mode";
import { useKioskState } from "./hooks/use-kiosk-state";
import { useIdle } from "./hooks/use-idle";
import { useSystemRefresh } from "./hooks/use-system-refresh";

function App() {
  const { activeNotification } = useNotifications();
  const { state } = useKioskState();
  const { isIdle } = useIdle();

  useSystemRefresh();

  const { mode } = useScreenMode({
    notification: activeNotification,
    isScreenOn: state.display.power === "on",
    isIdle,
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
