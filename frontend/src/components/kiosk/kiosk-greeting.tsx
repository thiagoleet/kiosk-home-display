import { useTranslation } from "../../hooks/use-translation";

function getGreetingKey() {
  const hour = new Date().getHours();

  if (hour < 12) {
    return "greeting.morning";
  }

  if (hour < 18) {
    return "greeting.afternoon";
  }

  return "greeting.evening";
}

export function KioskGreeting() {
  const { t } = useTranslation();

  return <p className="kiosk-greeting">{t(getGreetingKey())}</p>;
}
