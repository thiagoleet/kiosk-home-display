import { useEffect, useState } from "react";

function getCurrentTime() {
  return new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date());
}

export function KioskClock() {
  const [time, setTime] = useState(getCurrentTime);

  useEffect(() => {
    const interval = window.setInterval(() => {
      setTime(getCurrentTime());
    }, 1000);

    return () => {
      window.clearInterval(interval);
    };
  }, []);

  return (
    <time
      className="kiosk-clock"
      dateTime={new Date().toISOString()}
    >
      {time}
    </time>
  );
}
