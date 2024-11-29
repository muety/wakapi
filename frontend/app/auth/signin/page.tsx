import { Metadata } from "next";
import Link from "next/link";

import { UserAuthForm } from "@/components/user-auth-form";

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
    <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
      <div className="flex flex-col justify-center justify-items-center space-y-2 text-center align-middle">
        {/* <h1 className="text-2xl font-semibold tracking-tight">Welcome back</h1> */}
        <p className="text-sm text-muted-foreground">
          Enter your email to sign in to your account
        </p>
      </div>
      <UserAuthForm error={error} />
      {/* <Form /> */}
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
