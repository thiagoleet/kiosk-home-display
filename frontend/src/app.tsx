// Components
import { KioskLayout } from "./components/kiosk/kiosk-layout";
import { KioskScreen } from "./components/kiosk/kiosk-screen";

// Hooks

import { useNotifications } from "./hooks/use-notifications";
import { useScreenMode } from "./hooks/use-screen-mode";

function App() {
  const { activeNotification } = useNotifications();

  const { mode } = useScreenMode({
    notification: activeNotification,
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
