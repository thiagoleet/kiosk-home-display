import { ConnectionStatus } from "./components/connection-status";
import { DisplayStatus } from "./components/display-status";
import { useKioskState } from "./hooks/use-kiosk-state";
import { useWebSocket } from "./hooks/use-websocket";

const websocketUrl = "ws://localhost:8080/ws";

function App() {
  console.log("App render");

  const { state, handleMessage } = useKioskState();

  const { status } = useWebSocket({
    url: websocketUrl,
    onMessage: handleMessage,
  });

  return (
    <main>
      <h1>Kiosk Home Display</h1>

      <ConnectionStatus status={status} />

      <DisplayStatus display={state.display} />
    </main>
  );
}

export default App;
