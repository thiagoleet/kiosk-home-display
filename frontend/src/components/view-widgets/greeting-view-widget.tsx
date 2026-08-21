import { useTranslation } from "@/hooks/use-translation";
import { useGreetingKey } from "@/hooks/use-greeting";

export function GreetingViewWidget() {
  const { t } = useTranslation();
  const greetingKey = useGreetingKey();

  return <p className="kiosk-greeting">{t(greetingKey)}</p>;
}
