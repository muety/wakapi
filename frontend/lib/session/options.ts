import { SessionOptions } from "iron-session";

const { AUTH_SECRET } = process.env;

export interface SessionData {
  user: SessionUser;
  isLoggedIn: boolean;
  token: string;
}

export interface SessionUser {
  email: string;
  token: string;
  id: string;
  avatar: string;
  has_wakatime_integration: boolean;
}

export const defaultSession: SessionData = {
  user: {
    email: "",
    token: "",
    id: "",
    avatar: "",
    has_wakatime_integration: false,
  },
  isLoggedIn: false,
  token: "",
};

export const sessionOptions: SessionOptions = {
  password: AUTH_SECRET!,
  cookieName: "wakatimer-auth-session",
  cookieOptions: {
    // secure only works in `https` environments
    // if your localhost is not on `https`, then use: `secure: process.env.NODE_ENV === "production"`
    secure: true,
    // domain: "/",
  },
  ttl: 60 * 60 * 24, // 24 hours
};

export function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
