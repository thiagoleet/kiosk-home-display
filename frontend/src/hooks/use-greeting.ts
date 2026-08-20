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

export function useGreetingKey() {
  return getGreetingKey();
}
