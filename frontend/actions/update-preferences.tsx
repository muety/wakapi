"use server";

import { updateData } from "./api";

type Preference =
  | "hireable"
  | "show_email_in_public"
  | "public_leaderboard"
  | "heartbeats_timeout_sec";

export const saveProfile = async (data: Record<string, any>) => {
  const response = await updateData("/v1/profile", data);

  if (!response.success) {
    console.log("data", response.error);
    return { ok: false, data: { error: response.error } };
  }

  return { ok: true, data: response.data };
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
