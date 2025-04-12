import { getIronSession } from "iron-session";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { NextRequest } from "next/server";

import { postData } from "@/actions/api";
import { createIronSession } from "@/lib/server/auth";
import {
  defaultSession,
  SessionData,
  sessionOptions,
} from "@/lib/session/options";

export async function POST(request: NextRequest) {
  let requestData;

  try {
    requestData = await request.json();
  } catch (error) {
    return Response.json(
      {
        message: "Error parsing json request",
        error,
      },
      { status: 400 }
    );
  }

  const { email = "", password } = requestData as {
    email: string;
    password: string;
  };

  try {
    const apiResponse = await postData(
      "/v1/auth/login",
      { email, password },
      { skipAuth: true }
    );

    const json = apiResponse.data as {
      data: SessionData;
      status?: number;
    };

    if (!apiResponse.success) {
      return Response.json(json, { status: json.status || 500 });
    }

    const session = await createIronSession(json.data);

    return Response.json(session);
  } catch (error) {
    console.log("Error logging in", error);
    return Response.json(
      {
        message: "Error logging in",
        error,
      },
      { status: 500 }
    );
  }
}

// this is used for something pretty shitty. remove
export async function PUT(request: NextRequest) {
  const session = await getIronSession<SessionData>(cookies(), sessionOptions);

  if (session.isLoggedIn !== true) {
    return Response.json(defaultSession);
  }

  const requestData = await request.json();
  const { has_wakatime_integration = false } = requestData;

  session.user.has_wakatime_integration = has_wakatime_integration;
  await session.save();

  return Response.json(session);
}

// read session
export async function GET(request: NextRequest) {
  const session = await getIronSession<SessionData>(cookies(), sessionOptions);

  const action = new URL(request.url).searchParams.get("action");
  if (action === "logout") {
    session.destroy();
    return redirect("/");
  }

  if (session.isLoggedIn !== true) {
    return Response.json(defaultSession);
  }

  if (action === "user") {
    return Response.json(session.user);
  }

  return Response.json(session);
}

// logout
export async function DELETE() {
  const session = await getIronSession<SessionData>(cookies(), sessionOptions);

  session.destroy();
  return redirect("/");
}
