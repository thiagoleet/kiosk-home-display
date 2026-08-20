import type { Activity } from "../types/activity";

export async function getActivities(): Promise<Activity[]> {
  const response = await fetch("http://localhost:8080/api/activities");

  if (!response.ok) {
    throw new Error("Failed to fetch activities");
  }

  return response.json();
}
