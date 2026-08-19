import type { ReactNode } from "react";

type NotificationLayoutProps = {
  children: ReactNode;
};

export function NotificationLayout({ children }: NotificationLayoutProps) {
  return <section data-layout="notification">{children}</section>;
}
