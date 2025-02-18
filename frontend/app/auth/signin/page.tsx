import { Metadata } from "next";
import Link from "next/link";

import { EmailLoginForm } from "@/components/email-login-form";
import { OTPSignIn } from "@/components/otp-sign-in";
import { SocialLogin } from "@/components/social-login";

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
      <h4 className="font-medium pb-1 text-3xl mb-5">Login to wakana.</h4>
      <OTPSignIn />
      <SocialLogin />
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
