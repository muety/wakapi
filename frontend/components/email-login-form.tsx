// "use client";

// import { zodResolver } from "@hookform/resolvers/zod";
// import { useRouter } from "next/navigation";
// import * as React from "react";
// import { useFormState, useFormStatus } from "react-dom";
// import { useForm } from "react-hook-form";
// import * as z from "zod";

// import { loginAction, loginEmailAction } from "@/actions/auth";
// import { Icons } from "@/components/icons";
// import { buttonVariants } from "@/components/ui/button";
// import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";
// import { toast } from "@/components/ui/use-toast";
// import { startGithubLoginFlow } from "@/lib/oauth/github";
// import { cn } from "@/lib/utils";
// import { userNameSchema } from "@/lib/validations/user";

// import { Form } from "./ui/form";

// interface UserLoginFormProps extends React.HTMLAttributes<HTMLDivElement> {
//   loading?: boolean;
//   error?: string;
//   message?: string;
// }

// type FormData = z.infer<typeof userNameSchema>;

// export function EmailLoginForm({
//   className,
//   error,
//   message = "",
//   ...props
// }: UserLoginFormProps) {
//   const form = useForm<FormData>({
//     resolver: zodResolver(userNameSchema),
//   });
//   const [isGitHubLoading, setIsGitHubLoading] = React.useState<boolean>(false);
//   const router = useRouter();
//   const [state, formAction] = useFormState(loginEmailAction, {});
//   const errors = form.formState.errors;
//   const formRef = React.useRef<HTMLFormElement | null>(null);

//   if (error) {
//     toast({
//       title: "Error",
//       description: error,
//       variant: "destructive",
//     });
//   }

//   if (message) {
//     toast({
//       title: "Message",
//       description: message,
//       variant: "success",
//     });
//   }

//   const SubmitButton = ({ onClick }: { onClick: any }) => {
//     const { pending } = useFormStatus();
//     return (
//       <button
//         type="button"
//         className={cn(buttonVariants())}
//         onClick={onClick}
//         disabled={pending}
//       >
//         {pending && <Icons.spinner className="mr-2 size-4 animate-spin" />}
//         Continue
//       </button>
//     );
//   };

//   const loginHandler = async () => {
//     if (!(await form.trigger())) {
//       return toast({
//         title: "Invalid form",
//         description: "Please fill out all fields.",
//         variant: "destructive",
//       });
//     }

//     if (formRef.current) {
//       formRef.current.requestSubmit();
//     }
//   };

//   React.useEffect(() => {
//     const message = state.message;
//     if (state.message) {
//       toast({
//         title: message.title,
//         description: message.description,
//         variant: message.variant,
//       });
//     }
//   }, [state]);

//   return (
//     <Form {...form}>
//       <div className={cn("grid gap-6", className)} {...props}>
//         <form action={formAction} ref={formRef}>
//           <div className="grid gap-2">
//             <div className="grid gap-1">
//               <div className="my-2 grid">
//                 <Label className="sr-only my-1 text-left" htmlFor="email">
//                   Email
//                 </Label>
//                 <Input
//                   id="email"
//                   placeholder="Email"
//                   type="email"
//                   autoCapitalize="none"
//                   autoComplete="email"
//                   autoCorrect="off"
//                   {...form.register("email")}
//                 />
//                 {errors?.email && (
//                   <p className="px-1 text-xs text-red-600">
//                     {errors.email.message}
//                   </p>
//                 )}
//               </div>

//               {/* <div className="my-2 grid">
//                 <Label className="sr-only my-1 text-left" htmlFor="password">
//                   Password
//                 </Label>
//                 <Input
//                   id="password"
//                   placeholder="Password"
//                   type="password"
//                   autoCapitalize="none"
//                   autoComplete="off"
//                   autoCorrect="off"
//                   {...form.register("password")}
//                 />

//                 {errors?.password && (
//                   <p className="px-1 text-xs text-red-600">
//                     {errors.password.message}
//                   </p>
//                 )}
//               </div> */}
//             </div>
//             <SubmitButton onClick={loginHandler} />
//           </div>
//         </form>
//         <div className="relative">
//           <div className="absolute inset-0 flex items-center">
//             <span className="w-full border-t" />
//           </div>
//           <div className="relative flex justify-center text-xs uppercase">
//             <span className="bg-background px-2 text-muted-foreground">
//               Or continue with
//             </span>
//           </div>
//         </div>
//         <button
//           type="button"
//           className={cn(buttonVariants({ variant: "outline" }))}
//           onClick={() => {
//             setIsGitHubLoading(true);
//             // initiate github signin flow
//             router.push(startGithubLoginFlow());
//           }}
//           disabled={isGitHubLoading}
//         >
//           {isGitHubLoading ? (
//             <Icons.spinner className="mr-2 size-4 animate-spin" />
//           ) : (
//             <Icons.gitHub className="mr-2 size-4" />
//           )}{" "}
//           Github
//         </button>
//       </div>
//     </Form>
//   );
// }

"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { toast } from "@/components/hooks/use-toast";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";

const FormSchema = z.object({
  pin: z.string().min(6, {
    message: "Your one-time password must be 6 characters.",
  }),
});

export function InputOTPForm() {
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      pin: "",
    },
  });

  function onSubmit(data: z.infer<typeof FormSchema>) {
    toast({
      title: "You submitted the following values:",
      description: (
        <pre className="mt-2 w-[340px] rounded-md bg-slate-950 p-4">
          <code className="text-white">{JSON.stringify(data, null, 2)}</code>
        </pre>
      ),
    });
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="w-2/3 space-y-6">
        <FormField
          control={form.control}
          name="pin"
          render={({ field }) => (
            <FormItem>
              <FormLabel>One-Time Password</FormLabel>
              <FormControl>
                <InputOTP maxLength={6} {...field}>
                  <InputOTPGroup>
                    <InputOTPSlot index={0} />
                    <InputOTPSlot index={1} />
                    <InputOTPSlot index={2} />
                    <InputOTPSlot index={3} />
                    <InputOTPSlot index={4} />
                    <InputOTPSlot index={5} />
                  </InputOTPGroup>
                </InputOTP>
              </FormControl>
              <FormDescription>
                Please enter the one-time password sent to your phone.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit">Submit</Button>
      </form>
    </Form>
  );
}
