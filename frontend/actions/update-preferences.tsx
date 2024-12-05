"use server";

type Preference = "hireableBadge" | "displayEmail" | "displayCodeTime";

export async function updatePreference(preference: Preference, value: boolean) {
  // In a real application, you would update the user's preference in the database
  // For this example, we'll just log the change
  console.log(`Updating ${preference} to ${value}`);

  // Simulate a delay to show loading state
  await new Promise((resolve) => setTimeout(resolve, 500));

  // Return the new value to update the client
  return value;
}
