import useSWR from "swr";
import useSWRMutation from "swr/mutation";

import { fetchJson } from "../utils";
import { defaultSession, SessionData, SessionUser } from "./options";

const sessionApiRoute = "/api/session";
const signUpRoute = "/api/session/signup";

function doLogin(
  url: string,
  { arg }: { arg: { email: string; password: string } }
) {
  return fetchJson<SessionData>(url, {
    method: "POST",
    body: JSON.stringify(arg),
  });
}

function doSignup(
  url: string,
  { arg }: { arg: { email: string; password: string; password_repeat: string } }
) {
  return fetchJson<SessionData>(url, {
    method: "POST",
    body: JSON.stringify(arg),
  });
}

function updateSession(
  url: string,
  { arg }: { arg: { has_wakatime_integration: boolean } }
) {
  return fetchJson<SessionData>(url, {
    method: "PUT",
    body: JSON.stringify(arg),
  });
}

function doLogout(url: string) {
  return fetchJson<SessionData>(url, {
    method: "DELETE",
  });
}

export function useSessionUser() {
  return useSWR(`${sessionApiRoute}?action=user`, fetchJson<SessionUser>);
}

export function useClientSession() {
  return useSWR(sessionApiRoute, fetchJson<SessionData>, {
    fallbackData: defaultSession,
  });
}

export default function useSession() {
  const { data: session, isLoading } = useSWR(
    sessionApiRoute,
    fetchJson<SessionData>,
    {
      fallbackData: defaultSession,
    }
  );

  const { trigger: login, isMutating: isLoggingIn } = useSWRMutation(
    sessionApiRoute,
    doLogin,
    {
      revalidate: false,
    }
  );
  const { trigger: createAccount } = useSWRMutation(signUpRoute, doSignup, {
    revalidate: false,
  });
  const { trigger: modifySession } = useSWRMutation(
    sessionApiRoute,
    updateSession,
    {
      revalidate: true,
    }
  );
  const { trigger: logout, isMutating: isLoggingOut } = useSWRMutation(
    sessionApiRoute,
    doLogout
  );

  return {
    session,
    logout,
    login,
    createAccount,
    isLoading,
    modifySession,
    isLoggingIn,
    isLoggingOut,
  };
}
