"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import * as React from "react";
import { useForm } from "react-hook-form";
import * as z from "zod";

import { Icons } from "@/components/icons";
import { buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "@/components/ui/use-toast";
import useSession from "@/lib/session/use-session";
import { cn } from "@/lib/utils";
import { userSignupSchema } from "@/lib/validations/user";

interface UserAuthFormProps extends React.HTMLAttributes<HTMLDivElement> {
  loading?: boolean;
}

type FormData = z.infer<typeof userSignupSchema>;

export function UserSignUpAuthForm({ className, ...props }: UserAuthFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(userSignupSchema),
  });
  const [isLoading, setIsLoading] = React.useState<boolean>(false);
  const [isGitHubLoading, setIsGitHubLoading] = React.useState<boolean>(false);
  const { createAccount } = useSession();
  const router = useRouter();

  async function onSubmit(data: FormData) {
    setIsLoading(true);

    if (!data.email || !data.password) {
      return toast({
        title: "Invalid form",
        description: "Please fill out all fields.",
        variant: "destructive",
      });
    }

    const email = data.email.toLowerCase();
    const password = data.password;
    const password_repeat = data.password_repeat;

    const res = (await createAccount({
      email,
      password,
      password_repeat,
    })) as any;

    console.log("create account response", res);

    if (res.status > 202) {
      setIsLoading(false);
      return toast({
        title:
          "Error: account could not be created at the moment. Try again later",
        description: res.message,
        variant: "destructive",
      });
    }

    toast({
      title: "Account created",
      description:
        "Your account has been created. You can now login. Happy hacking.",
      variant: "success",
    });
    router.push("/dashboard");
  }

  return (
    <div className={cn("grid gap-6", className)} {...props}>
      <form onSubmit={handleSubmit(onSubmit)}>
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
                disabled={isLoading}
                {...register("email")}
              />
              {errors?.email && (
                <p className="px-1 text-xs text-red-600">
                  {errors.email.message}
                </p>
              )}
            </div>

            <div className="my-2 grid">
              <Label className="sr-only my-1 text-left" htmlFor="password">
                Password
              </Label>
              <Input
                id="password"
                placeholder="Password"
                type="password"
                autoCapitalize="none"
                autoComplete="off"
                autoCorrect="off"
                disabled={isLoading}
                {...register("password")}
              />
              {errors?.password && (
                <p className="px-1 text-xs text-red-600">
                  {errors.password.message}
                </p>
              )}
            </div>

            <div className="my-1 grid">
              <Label
                className="sr-only my-2 text-left"
                htmlFor="password_repeat"
              >
                Repeat Password
              </Label>
              <Input
                id="password_repeat"
                placeholder="Repeat Password"
                type="password"
                autoCapitalize="none"
                autoComplete="off"
                autoCorrect="off"
                disabled={isLoading}
                {...register("password_repeat")}
              />
              {errors?.password_repeat && (
                <p className="px-1 text-xs text-red-600">
                  {errors.password_repeat.message}
                </p>
              )}
            </div>
          </div>
          <button className={cn(buttonVariants())} disabled={isLoading}>
            {isLoading && (
              <Icons.spinner className="mr-2 size-4 animate-spin" />
            )}
            Sign Up
          </button>
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
        }}
        disabled={isLoading || isGitHubLoading}
      >
        {isGitHubLoading ? (
          <Icons.spinner className="mr-2 size-4 animate-spin" />
        ) : (
          <Icons.gitHub className="mr-2 size-4" />
        )}{" "}
        Github
      </button>
    </div>
  );
}
