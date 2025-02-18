import { Metadata } from "next";
import Link from "next/link";

import { UserLoginForm } from "@/components/user-auth-form";

export const metadata: Metadata = {
  title: "Login",
  description: "Login to your account",
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const { error, message } = searchParams;

  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight mb-2">
        Login to your account
      </h1>
      <UserLoginForm error={error} message={message} />
      <p className="px-8 my-2 text-center text-sm text-muted-foreground">
        <Link
          href="/auth/signup"
          className="hover:text-brand underline underline-offset-4"
        >
          Don&apos;t have an account? Sign Up
        </Link>
      </p>
      <p className="px-8 my-2 text-center text-sm text-muted-foreground">
        <Link
          href="/auth/forgot-password"
          className="hover:text-brand underline underline-offset-4"
        >
          Forgot your password?
        </Link>
      </p>
    </div>
  );
}
