"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button, buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Form, FormControl, FormField, FormItem } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/use-toast";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { initiateOTPLoginAction } from "@/actions/auth";
import { useFormState, useFormStatus } from "react-dom";
import { Icons } from "./icons";
import React from "react";
import { PKCEGenerator, PKCEResult } from "@/lib/oauth/pkce";

const formSchema = z.object({
  email: z.string().email(),
});

type Props = {
  className?: string;
};

const SubmitButton = ({ onClick }: { onClick: any }) => {
  const { pending } = useFormStatus();
  return (
    //     <Button
    //     type="submit"
    //     className="active:scale-[0.98] bg-primary px-6 py-4 text-secondary font-medium flex space-x-2 h-[40px] w-full"
    //   >
    //     {isLoading ? (
    //       <Loader2 className="h-4 w-4 animate-spin" />
    //     ) : (
    //       <span>Continue</span>
    //     )}
    //   </Button>
    <button
      type="submit"
      className={cn(buttonVariants())}
      onClick={onClick}
      disabled={pending}
    >
      {pending && <Icons.spinner className="mr-2 size-4 animate-spin" />}
      Continue
    </button>
  );
};

// initiateOTPLoginAction

export function OTPSignIn({ className }: Props) {
  const [state, formAction] = useFormState(initiateOTPLoginAction, {
    message: null,
  });
  const [isSent, setSent] = useState(false);
  const [email, setEmail] = useState<string>();
  const formRef = React.useRef<HTMLFormElement | null>(null);
  const [challengeVerifier, setChallengeVerifier] = useState<string>();
  const [pkce, setPKCE] = useState<PKCEResult>();

  useEffect(() => {
    const pkceFactory = new PKCEGenerator();
    pkceFactory.generatePKCE().then((result) => setPKCE(result));
  }, []);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: "",
    },
  });

  React.useEffect(() => {
    const message = state.message;
    if (state.message) {
      toast({
        title: message.title,
        description: message.description,
        variant: message.variant,
      });
    }

    if (state.success) {
      setSent(true);
      setEmail(form.getValues("email"));
      setChallengeVerifier(state.challengeVerifier);
    }
    console.log("state", state);
  }, [state]);

  const verifyOtp = {
    status: "idle",
  };

  const loginHandler = async () => {
    if (!(await form.trigger())) {
      return toast({
        title: "Invalid form",
        description: "Please fill out all fields.",
        variant: "destructive",
      });
    }

    if (formRef.current) {
      formRef.current.requestSubmit();
    }
  };

  async function onComplete(token: string) {
    if (!email) return;
    console.log("token", { token, email, challengeVerifier });
    // verifyOtp.execute({
    //   token,
    //   email,
    // });
  }

  if (isSent) {
    return (
      <div className={cn("flex flex-col space-y-4 items-center", className)}>
        <InputOTP
          maxLength={6}
          autoFocus
          onComplete={onComplete}
          disabled={verifyOtp.status === "executing"}
          render={({ slots }) => (
            <InputOTPGroup>
              {slots.map((slot, index) => (
                <InputOTPSlot
                  key={index.toString()}
                  {...slot}
                  className="w-[62px] h-[62px]"
                />
              ))}
            </InputOTPGroup>
          )}
        />

        <div className="flex space-x-2">
          <span className="text-sm text-[#878787]">
            Didn't receive the email?
          </span>
          <button
            onClick={() => setSent(false)}
            type="button"
            className="text-sm text-primary underline font-medium"
          >
            Resend code
          </button>
        </div>
      </div>
    );
  }

  return (
    <Form {...form}>
      <form action={formAction}>
        <div className={cn("flex flex-col space-y-4", className)}>
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input
                    placeholder="Enter email address"
                    {...field}
                    autoCapitalize="false"
                    autoCorrect="false"
                    spellCheck="false"
                  />
                </FormControl>
              </FormItem>
            )}
          />
          <input
            type="hidden"
            name="code_challenge"
            value={pkce?.codeChallenge}
          />
          <input
            type="hidden"
            name="challenge_method"
            value={pkce?.codeChallengeMethod}
          />
          <SubmitButton onClick={loginHandler} />
        </div>
      </form>
    </Form>
  );
}
