"use server";

import { redirect } from "next/navigation";
import { SessionData } from "@/lib/session/options";
import { createIronSession } from "@/lib/server/auth";
import { userNameSchema } from "@/lib/validations/user";

const { NEXT_PUBLIC_API_URL } = process.env;

export async function loginAction(_: any, formData: FormData): Promise<any> {
  const validatedFields = userNameSchema.safeParse({
    email: formData.get("email"),
    password: formData.get("password"),
  });

  if (!validatedFields.success) {
    return {
      message: {
        description: "Email and password are required",
        title: "Error",
        variant: "destructive",
      },
    };
  }

  return processLogin(validatedFields.data);
}

export async function processLogin(credentials: {
  email: string;
  password: string;
}) {
  let redirectPath = null;
  try {
    const apiResponse = await fetch(`${NEXT_PUBLIC_API_URL}/api/auth/login`, {
      method: "POST",
      headers: {
        accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(credentials),
    });

    const json = (await apiResponse.json()) as {
      data: SessionData;
      status?: number;
      message?: string;
    };

    if (apiResponse.status > 202) {
      return {
        message: {
          description: json.message || "Unexpected error logging in",
          title: "Login Error!",
          variant: "default",
        },
      };
    }

    await createIronSession(json.data);
    redirectPath = "/dashboard";
  } catch (error) {
    console.log("Error logging in", error);
    return {
      message: {
        description: "Unexpected error logging in",
        title: "Login Error!",
        variant: "default",
      },
    };
  } finally {
    redirectPath && redirect(redirectPath);
    // redirect apparently throws an error so shouldn't be called inside a try catch
    // https://nextjs.org/docs/app/building-your-application/routing/redirecting#redirect-function
  }
}
