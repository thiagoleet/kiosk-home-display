import type { WebSocketStatus } from "../hooks/use-websocket";
import { useTranslation } from "../hooks/use-translation";

type ConnectionStatusProps = {
  status: WebSocketStatus;
};

export function ConnectionStatus({ status }: ConnectionStatusProps) {
  const { t } = useTranslation();

  return <div>{t("connection.status", { status })}</div>;
}
