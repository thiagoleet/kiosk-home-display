export type NotificationLevel = "info" | "success" | "warning" | "error";

export type Notification = {
  title: string;
  message: string;
  level: NotificationLevel;
};
