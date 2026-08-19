export type NotificationContext = "printer" | "system" | "network";

export type NotificationLevel = "info" | "success" | "warning" | "error";

export type Notification = {
  id: string;
  context: NotificationContext;
  title: string;
  message: string;
  level: NotificationLevel;
  duration: number;
};
