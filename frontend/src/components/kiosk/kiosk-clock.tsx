import { useEffect, useState } from "react";
import { useTranslation } from "../../hooks/use-translation";

function getCurrentTime(locale: string) {
  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date());
}

export function KioskClock() {
  const { locale } = useTranslation();
  const [time, setTime] = useState(() => getCurrentTime(locale));

  useEffect(() => {
    const updateTime = () => {
      setTime(getCurrentTime(locale));
    };

    const initialTimer = window.setTimeout(updateTime, 0);
    const interval = window.setInterval(updateTime, 1000);

    return () => {
      window.clearTimeout(initialTimer);
      window.clearInterval(interval);
    };
  }, [locale]);

  return (
    <time
      className="kiosk-clock"
      dateTime={new Date().toISOString()}
    >
      {time}
    </time>
  );
}
