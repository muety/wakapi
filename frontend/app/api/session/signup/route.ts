import { NextRequest } from "next/server";

import { createIronSession } from "@/lib/server/auth";
import { SessionData } from "@/lib/session/options";

const { NEXT_PUBLIC_API_URL } = process.env;

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

  const {
    email = "",
    password,
    password_repeat,
  } = requestData as {
    email: string;
    password: string;
    password_repeat: string;
  };

  if (password !== password_repeat) {
    return Response.json(
      {
        message: "Password mismatch",
      },
      { status: 400 }
    );
  }

  try {
    const apiResponse = await fetch(`${NEXT_PUBLIC_API_URL}/api/auth/signup`, {
      method: "POST",
      headers: {
        accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({
        email,
        password,
        password_repeat,
      }),
    });

    if (apiResponse.status > 202 || !apiResponse.status) {
      return Response.json(apiResponse.body, {
        status: apiResponse.status || 500,
      });
    }
    const json = (await apiResponse.json()) as { data: SessionData };

    const session = await createIronSession(json.data);

    session.isLoggedIn = true;
    session.user = json.data.user;
    session.token = json.data.token;
    await session.save();

    return Response.json(session);
  } catch (error) {
    console.log("Error signing up", error);
    return Response.json(
      {
        message: "Error logging in",
        error,
      },
      { status: 400 }
    );
  }
}
