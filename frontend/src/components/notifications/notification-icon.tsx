import {
  BellOutlined,
  PrinterOutlined,
  SettingOutlined,
  WifiOutlined,
} from "@ant-design/icons";

import type { NotificationContext } from "../../types/notification";

type NotificationIconProps = {
  context: NotificationContext;
};

export function NotificationIcon({ context }: NotificationIconProps) {
  switch (context) {
    case "printer":
      return <PrinterOutlined />;

    case "network":
      return <WifiOutlined />;

    case "system":
      return <SettingOutlined />;

    default:
      return <BellOutlined />;
  }
}
