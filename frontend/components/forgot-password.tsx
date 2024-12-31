"use client";

import * as z from "zod";
import * as React from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useFormState, useFormStatus } from "react-dom";

import { cn } from "@/lib/utils";
import { Icons } from "@/components/icons";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "@/components/ui/use-toast";
import { forgotPasswordAction } from "@/actions/auth";
import { buttonVariants } from "@/components/ui/button";
import { startGithubLoginFlow } from "@/lib/oauth/github";
import { forgotPasswordSchema } from "@/lib/validations/user";

import { Form } from "./ui/form";

interface ForgotPasswordFormProps extends React.HTMLAttributes<HTMLDivElement> {
  loading?: boolean;
  error?: string;
}

type FormData = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordForm({
  className,
  error,
  ...props
}: ForgotPasswordFormProps) {
  const form = useForm<FormData>({
    resolver: zodResolver(forgotPasswordSchema),
  });
  const [isGitHubLoading, setIsGitHubLoading] = React.useState<boolean>(false);
  const router = useRouter();
  const [state, formAction] = useFormState(forgotPasswordAction, {});
  const errors = form.formState.errors;
  const formRef = React.useRef<HTMLFormElement | null>(null);

  if (error) {
    toast({
      title: "Error",
      description: error,
      variant: "destructive",
    });
  }

  const SubmitButton = ({ onClick }: { onClick: any }) => {
    const { pending } = useFormStatus();
    return (
      <button
        type="button"
        className={cn(buttonVariants())}
        onClick={onClick}
        disabled={pending}
      >
        {pending && <Icons.spinner className="mr-2 size-4 animate-spin" />}
        Sign In
      </button>
    );
  };

  const formHandler = async () => {
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

  React.useEffect(() => {
    const message = state.message;
    if (state.message) {
      toast({
        title: message.title,
        description: message.description,
        variant: message.variant,
      });
    }
  }, [state]);

  return (
    <Form {...form}>
      <div className={cn("grid gap-6", className)} {...props}>
        <form action={formAction} ref={formRef}>
          <div className="grid gap-2">
            <div className="grid gap-1">
              <div className="my-2 grid">
                <Label className="sr-only my-1 text-left" htmlFor="email">
                  Email
                </Label>
                <Input
                  id="email"
                  placeholder="Email"
                  type="email"
                  autoCapitalize="none"
                  autoComplete="email"
                  autoCorrect="off"
                  {...form.register("email")}
                />
                {errors?.email && (
                  <p className="px-1 text-xs text-red-600">
                    {errors.email.message}
                  </p>
                )}
              </div>
            </div>
            <SubmitButton onClick={formHandler} />
          </div>
        </form>
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">
              Or continue with
            </span>
          </div>
        </div>
        <button
          type="button"
          className={cn(buttonVariants({ variant: "outline" }))}
          onClick={() => {
            setIsGitHubLoading(true);
            // initiate github signin flow
            router.push(startGithubLoginFlow());
          }}
          disabled={isGitHubLoading}
        >
          {isGitHubLoading ? (
            <Icons.spinner className="mr-2 size-4 animate-spin" />
          ) : (
            <Icons.gitHub className="mr-2 size-4" />
          )}{" "}
          Github
        </button>
      </div>
    </Form>
  );
}
