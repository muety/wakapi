import { Metadata } from "next";
import Link from "next/link";

import { ForgotPasswordForm } from "@/components/forgot-password";

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
      <h1 className="text-2xl font-semibold tracking-tight">
        Reset Your Password
      </h1>
      <div className="flex flex-col justify-center justify-items-center space-y-2 text-center align-middle">
        <p className="text-sm text-muted-foreground">
          Enter your email to receive a password reset link
        </p>
      </div>
      <ForgotPasswordForm error={error} />
      <p className="px-8 my-2 text-center text-sm text-muted-foreground">
        <Link
          href="/auth/signin"
          className="hover:text-brand underline underline-offset-4"
        >
          Remembered your password? Sign In
        </Link>
      </p>
    </div>
  );
}
