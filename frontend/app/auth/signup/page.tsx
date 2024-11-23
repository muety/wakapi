import Link from "next/link";

import { Metadata } from "next";
import { UserSignUpAuthForm } from "@/components/user-signup-form";

export const metadata: Metadata = {
  title: "Login",
  description: "Login to your account",
};

export default async function LoginPage() {
  return (
    <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
      <div className="flex flex-col justify-center align-middle justify-items-center space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Create a New Account
        </h1>
        <p className="text-sm text-muted-foreground">
          Enter your email and password to get started
        </p>
      </div>
      <UserSignUpAuthForm />
      {/* <Form /> */}
      <p className="px-8 text-center text-sm text-muted-foreground">
        <Link
          href="/auth/signin"
          className="hover:text-brand underline underline-offset-4"
        >
          Already have an account? Sign In
        </Link>
      </p>
    </div>
  );
}
