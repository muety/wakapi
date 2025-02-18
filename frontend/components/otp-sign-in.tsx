"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import React from "react";
import { useFormState, useFormStatus } from "react-dom";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { initiateOTPLoginAction, processLoginWithOTP } from "@/actions/auth";
import { buttonVariants } from "@/components/ui/button";
import { Form, FormControl, FormField, FormItem } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";
import { toast } from "@/components/ui/use-toast";
import { PKCEGenerator, PKCEResult } from "@/lib/oauth/pkce";
import { cn } from "@/lib/utils";

import { Icons } from "./icons";

const formSchema = z.object({
  email: z.string().email(),
});

type Props = {
  className?: string;
};

const SubmitButton = ({ onClick }: { onClick: any }) => {
  const { pending } = useFormStatus();
  return (
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

export function OTPSignIn({ className }: Props) {
  const [state, formAction] = useFormState(initiateOTPLoginAction, {
    message: null,
  });
  const [isSent, setSent] = useState(false);
  const [email, setEmail] = useState<string>();
  const formRef = React.useRef<HTMLFormElement | null>(null);
  const [pkce, setPKCE] = useState<PKCEResult>();
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const pkceFactory = new PKCEGenerator();
    pkceFactory.generatePKCE().then((result) => {
      setPKCE(result);
    });
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
    }
    console.log("state", state);
  }, [state, form]);

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

  async function onComplete(otp: string) {
    try {
      setIsLoading(true);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
    if (!email || !pkce?.code_verifier) {
      return toast({
        title: "Invalid OTP",
        description: "Please check your OTP and try again.",
        variant: "destructive",
      });
    }

    const response = await processLoginWithOTP({
      email,
      otp,
      code_verifier: pkce?.code_verifier || "",
    });

    if (response && response.message) {
      const { message } = response;
      toast({
        title: message.title,
        description: message.description,
        variant: "destructive",
      });
    }
  }

  if (isSent) {
    return (
      <div className={cn("flex flex-col space-y-4 items-center", className)}>
        <InputOTP
          maxLength={6}
          autoFocus
          onComplete={onComplete}
          disabled={isLoading}
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
            value={pkce?.code_challenge}
          />
          <input
            type="hidden"
            name="challenge_method"
            value={pkce?.challenge_method}
          />
          <SubmitButton onClick={loginHandler} />
        </div>
      </form>
    </Form>
  );
}
