export type WebSocketMessageType =
  | "state.snapshot"
  | "display.state_changed"
  | "idle.timeout"
  | "schedule.on"
  | "schedule.off"
  | "printer.started"
  | "printer.completed"
  | "notification";

export type WebSocketMessage<T = unknown> = {
  type: WebSocketMessageType;
  data?: T;
};
