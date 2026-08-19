export type NotificationLevel = "info" | "success" | "warning" | "error";

export type Notification = {
  id: string;
  title: string;
  message: string;
  level: NotificationLevel;
  duration: number;
};
