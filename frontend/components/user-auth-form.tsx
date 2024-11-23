"use client";

import * as z from "zod";
import * as React from "react";

import { cn } from "@/lib/utils";
import { Form } from "./ui/form";
import { useForm } from "react-hook-form";
import { Icons } from "@/components/icons";
import { useRouter } from "next/navigation";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/use-toast";
import { loginAction } from "@/actions/auth";
import { zodResolver } from "@hookform/resolvers/zod";
import { userNameSchema } from "@/lib/validations/user";
import { buttonVariants } from "@/components/ui/button";
import { useFormState, useFormStatus } from "react-dom";
import { startGithubLoginFlow } from "@/lib/oauth/github";

interface UserAuthFormProps extends React.HTMLAttributes<HTMLDivElement> {
  loading?: boolean;
  error?: string;
}

type FormData = z.infer<typeof userNameSchema>;

export function UserAuthForm({
  className,
  error,
  ...props
}: UserAuthFormProps) {
  const form = useForm<FormData>({
    resolver: zodResolver(userNameSchema),
  });
  const [isGitHubLoading, setIsGitHubLoading] = React.useState<boolean>(false);
  const router = useRouter();
  const [state, formAction] = useFormState(loginAction, {});
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
        {pending && <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />}
        Sign In
      </button>
    );
  };

  const loginHandler = async (event: React.FormEvent<HTMLFormElement>) => {
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
              <div className="grid my-2">
                <Label className="text-left sr-only my-1" htmlFor="email">
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

              <div className="grid my-2">
                <Label className="text-left sr-only my-1" htmlFor="password">
                  Password
                </Label>
                <Input
                  id="password"
                  placeholder="Password"
                  type="password"
                  autoCapitalize="none"
                  autoComplete="off"
                  autoCorrect="off"
                  {...form.register("password")}
                />

                {errors?.password && (
                  <p className="px-1 text-xs text-red-600">
                    {errors.password.message}
                  </p>
                )}
              </div>
            </div>
            <SubmitButton onClick={loginHandler} />
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
            <Icons.spinner className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Icons.gitHub className="mr-2 h-4 w-4" />
          )}{" "}
          Github
        </button>
      </div>
    </Form>
  );
}
