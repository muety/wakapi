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
  const error = searchParams.error;

  return (
    <div>
      <div className="flex flex-col justify-center justify-items-center space-y-2 text-center align-middle">
        <p className="text-sm text-muted-foreground">
          Enter your email to sign in to your account
        </p>
      </div>
      <UserLoginForm error={error} />
      <p className="px-8 text-center text-sm text-muted-foreground">
        <Link
          href="/auth/signup"
          className="hover:text-brand underline underline-offset-4"
        >
          Don&apos;t have an account? Sign Up
        </Link>
      </p>
    </div>
  );
}
