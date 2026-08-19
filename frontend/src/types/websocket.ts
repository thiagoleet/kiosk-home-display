export type WebSocketMessageType =
  | "state.snapshot"
  | "display.state_changed"
  | "idle.timeout"
  | "schedule.on"
  | "schedule.off"
  | "printer.started"
  | "printer.completed"
  | "notification"
  | "activity";

export type WebSocketMessage<T = unknown> = {
  type: WebSocketMessageType;
  data?: T;
};

export type WebSocketStatus = "connecting" | "connected" | "disconnected";
