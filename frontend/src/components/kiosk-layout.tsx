import type { ReactNode } from "react";

import { HomeLayout } from "./layouts/home-layout";
import { NotificationLayout } from "./layouts/notification-layout";

import type { ScreenMode } from "../types/screen";

type KioskLayoutProps = {
  mode: ScreenMode;
  children: ReactNode;
};

export function KioskLayout({ mode, children }: KioskLayoutProps) {
  if (mode === "notification") {
    return <NotificationLayout>{children}</NotificationLayout>;
  }

  return <HomeLayout>{children}</HomeLayout>;
}
