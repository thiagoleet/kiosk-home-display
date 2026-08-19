function getGreeting() {
  const hour = new Date().getHours();

  if (hour < 12) {
    return "Good morning";
  }

  if (hour < 18) {
    return "Good afternoon";
  }

  return "Good evening";
}

export function KioskGreeting() {
  return <p className="kiosk-greeting">{getGreeting()}</p>;
}
