import { Metadata } from "next";

import { OTPSignIn } from "@/components/otp-sign-in";
import { SocialLogin } from "@/components/social-login";

export const metadata: Metadata = {
  title: "Login",
  description: "Login to your account",
};

export default async function LoginPage() {
  return (
    <div>
      <h1 className="font-medium pb-3 text-3xl">Login to wakana.</h1>
      <OTPSignIn />
      <SocialLogin />
      <p className="text-xs text-[#878787]">
        By clicking continue, you acknowledge that you have read and agree to
        Wakana's{" "}
        <a href="https://wakana.io/terms" className="underline">
          Terms of Service
        </a>{" "}
        and{" "}
        <a href="https://midday.ai/policy" className="underline">
          Privacy Policy
        </a>
        .
      </p>
    </div>
  );
}
