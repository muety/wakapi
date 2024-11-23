import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";

export async function mutateData(
  url: string,
  {
    arg: { method, body = {}, token },
  }: {
    arg: {
      method: "POST" | "PUT" | "DELETE";
      body?: Record<string, any>;
      token: string; // Accept token as a parameter
    };
  }
) {
  const bodyPayload: any =
    body && Object.keys(body).length > 0 ? { body: JSON.stringify(body) } : {};
  const res = await fetch(`${NEXT_PUBLIC_API_URL}/api/${url}`, {
    method,
    headers: {
      accept: "application/json",
      "content-type": "application/json",
      token, // Use passed token
    },
    ...bodyPayload,
  });
  return res.json();
}
