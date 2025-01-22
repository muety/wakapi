"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

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
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "@/components/ui/use-toast";
import { profileFormSchema, type ProfileFormValues } from "@/lib/schema";

export function ProfileForm() {
  // on;y set value for fields if the exist. Otherwise react0hoo-forms behaves weirdly
  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      // name: undefined,
      // username: undefined,
      // bio: undefined,
      // github: undefined,
      // twitter: undefined,
      // linkedin: undefined,
    },
  });

  // we need to fetch the users profile information
  // and use it to set default values

  // const form = useForm<ProfileFormValues>({
  //   resolver: zodResolver(profileFormSchema),
  //   mode: "onSubmit", // Ensures validation occurs on submit
  // });

  function onSubmit(data: ProfileFormValues) {
    // make the call to save these sorry ...
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
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="Your name" {...field} />
              </FormControl>
              <FormDescription>
                This is your public display name.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder="Your username" {...field} />
              </FormControl>
              <FormDescription>This is your public username. </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="bio"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Bio</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Tell us a little bit about yourself"
                  className="resize-none"
                  {...field}
                />
              </FormControl>
              {/* <FormDescription>
                You can <span>@mention</span> other users and organizations.
              </FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
        {/* <FormField
          control={form.control}
          name="key_stroke_timeout"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Keystroke Timeout (ms)</FormLabel>
              <FormControl>
                <Input placeholder="Keystroke timeout" {...field} />
              </FormControl>
              <FormDescription>
                Minutes between keystrokes before timeout is stopped.{" "}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        /> */}
        <FormField
          control={form.control}
          name="github"
          render={({ field }) => (
            <FormItem>
              <FormLabel>GitHub</FormLabel>
              <FormControl>
                <div className="flex">
                  <span className="inline-flex items-center rounded-l-md border border-r-0 px-3 text-sm text-gray-500">
                    hhtps://github.com/
                  </span>
                  <Input className="rounded-l-none" {...field} />
                </div>
              </FormControl>
              {/* <FormDescription>Enter your GitHub username.</FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="twitter"
          render={({ field }) => (
            <FormItem>
              <FormLabel>X (Twitter)</FormLabel>
              <FormControl>
                <div className="flex">
                  <span className="inline-flex items-center rounded-l-md border border-r-0 px-3 text-sm text-gray-500">
                    https://x.com/@
                  </span>
                  <Input className="rounded-l-none" {...field} />
                </div>
              </FormControl>
              {/* <FormDescription>
                Enter your X (Twitter) username.
              </FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="linkedin"
          render={({ field }) => (
            <FormItem>
              <FormLabel>LinkedIn</FormLabel>
              <FormControl>
                <div className="flex">
                  <span className="inline-flex items-center rounded-l-md border border-r-0 px-3 text-sm text-gray-500">
                    https://linkedin.com/in/
                  </span>
                  <Input className="rounded-l-none" {...field} />
                </div>
              </FormControl>
              {/* <FormDescription>Enter your LinkedIn username.</FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Update profile</Button>
      </form>
    </Form>
  );
}
