import type { NotificationContext } from "./notification";

export type Activity = {
  id: string;
  context: NotificationContext;
  type: string;
  title: string;
  description?: string;
  timestamp: string;
};
