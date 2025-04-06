"use server";

import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";

import { getSession } from "./session";

type Preference =
  | "hireable"
  | "show_email_in_public"
  | "public_leaderboard"
  | "heartbeats_timeout_sec";

export const saveProfile = async (data: Record<string, any>) => {
  const session = await getSession();
  const response = await fetch(`${NEXT_PUBLIC_API_URL}/api/v1/profile`, {
    method: "PUT",
    headers: {
      "content-type": "application/json",
      accept: "application/json",
      token: `${session.token}`,
    },
    body: JSON.stringify(data),
  });

  const response_data = await response.json();

  if (!response.ok) {
    console.log("data", response_data);
    return { ok: false, data: response_data };
  }

  return { ok: true, data: response_data };
};

export async function updatePreference(
  preference: Preference,
  value: boolean | string | number
) {
  // In a real application, you would update the user's preference in the database
  // For this example, we'll just log the change
  console.log(`Updating ${preference} to ${value}`);

  // Simulate a delay to show loading state
  await saveProfile({ [preference]: value });

  // Return the new value to update the client
  return value;
}
