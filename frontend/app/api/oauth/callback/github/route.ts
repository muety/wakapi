import { redirect } from "next/navigation";
import { NextRequest } from "next/server";

import { postData } from "@/actions/api";
import { getGithubConfig } from "@/lib/oauth/github";
import { createIronSession } from "@/lib/server/auth";
import { SessionData } from "@/lib/session/options";

// validates state and code from github oauth
export async function GET(request: NextRequest) {
  let redirectPath = "";
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
      throw new Error("state is invalid");
    }
    await handleGithubOauth(code);
    redirectPath = "/dashboard";
  } catch (error) {
    console.log("error logging in with github", error);
    const error_payload = {
      error:
        "An unexpected error occurred while logging in using github. Try again later. If this persists, contact support",
    };
    return redirect(`/login?${new URLSearchParams(error_payload).toString()}`);
  } finally {
    if (redirectPath) {
      redirect(redirectPath);
    }
  }
}

// throws
async function handleGithubOauth(code: string) {
  const apiResponse = await postData(
    "/v1/auth/oauth/github",
    { code },
    { skipAuth: true }
  );

  console.log("handleGithubOauth", apiResponse);

  const payload = apiResponse.data as {
    data: SessionData;
    status?: number;
  };

  if (!apiResponse.success) {
    throw new Error("Error logging in");
  }

  await createIronSession(payload.data);
}
