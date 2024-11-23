import { cookies } from "next/headers";
import { getIronSession } from "iron-session";
import { SessionData, sessionOptions } from "@/lib/session/options";

export async function createIronSession(data: SessionData) {
  const session = await getIronSession<SessionData>(cookies(), {
    ...sessionOptions,
  });
  session.isLoggedIn = true;
  session.user = data.user;
  session.token = data.token;
  await session.save();
  return session;
}
