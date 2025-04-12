import { NextRequest } from "next/server";

import { postData } from "@/actions/api";
import { createIronSession } from "@/lib/server/auth";
import { SessionData } from "@/lib/session/options";

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
    const apiResponse = await postData("/v1/auth/signup", {
      email,
      password,
      password_repeat,
    });

    if (!apiResponse.success) {
      return Response.json(apiResponse.data, {
        status: 500,
      });
    }
    const json = apiResponse.data as { data: SessionData };
    console.log("json", json);
    // broken until manual testing

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
