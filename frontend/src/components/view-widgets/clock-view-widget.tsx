import { useTranslation } from "@/hooks/use-translation";
import { useClock } from "@/hooks/use-clock";

export function ClockViewWidget() {
  const { locale } = useTranslation();
  const time = useClock(locale);

  return (
    <time
      className="kiosk-clock"
      dateTime={new Date().toISOString()}
    >
      {time}
    </time>
  );
}
