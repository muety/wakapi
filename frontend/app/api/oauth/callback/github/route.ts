import { redirect } from "next/navigation";
import { NextRequest } from "next/server";

import { getGithubConfig } from "@/lib/oauth/github";
import { createIronSession } from "@/lib/server/auth";
import { SessionData } from "@/lib/session/options";

const { NEXT_PUBLIC_API_URL } = process.env;

// validates state and code from github oauth
export async function GET(request: NextRequest) {
  try {
    const searchParams = new URL(request.url).searchParams;
    const state = searchParams.get("state");
    const code = searchParams.get("code");
    const config = getGithubConfig();

    if (!state || !code) {
      throw new Error("state and code are required");
    }
    // validate state
    const parsedState = JSON.parse(
      Buffer.from(state || "", "base64").toString()
    );
    const { redirectUri, scope, clientId } = parsedState;
    if (
      config.redirectUri !== redirectUri ||
      config.clientId !== clientId ||
      config.scope !== scope
    ) {
      console.log("INVALID STATE");
      throw new Error("state is invalid");
    }
    await handleGithubOauth(code);
    console.log("GITHUB OAUTH HANDLED");
  } catch (error) {
    console.log("error", error);
    const error_payload = {
      error:
        "An unexpected error occurred while logging in using github. Try again later. If this persists, contact support",
    };
    return redirect(
      `/auth/signin?${new URLSearchParams(error_payload).toString()}`
    );
  } finally {
    return redirect(`/dashboard`);
  }
}

// throws
async function handleGithubOauth(code: string) {
  console.log("Handing github oauth", new Date());
  const apiResponse = await fetch(
    `${NEXT_PUBLIC_API_URL}/api/auth/oauth/github`,
    {
      method: "POST",
      headers: {
        accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({ code }),
    }
  );

  const json = (await apiResponse.json()) as {
    data: SessionData;
    status?: number;
  };

  if (apiResponse.status > 202) {
    throw new Error("Error logging in");
  }

  const session = await createIronSession(json.data);

  return Response.json(session);
}
