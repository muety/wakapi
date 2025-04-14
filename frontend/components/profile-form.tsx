"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import React from "react";
import { useForm } from "react-hook-form";

import { saveProfile } from "@/actions/update-preferences";
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
import { UserProfile } from "@/lib/types";

import { Icons } from "./icons";

export function ProfileForm({ user }: { user: UserProfile }) {
  const [loading, setLoading] = React.useState(false);

  // only set value for fields if the exist. Otherwise react-hool-forms behaves weirdly
  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      name: user?.name || undefined,
      username: user?.username || undefined,
      bio: user?.bio || undefined,
      github_handle: user?.github_handle || undefined,
      twitter_handle: user?.twitter_handle || undefined,
      linked_in_handle: user?.linked_in_handle || undefined,
    },
  });

  const onSubmit = async (data: ProfileFormValues) => {
    setLoading(true);
    try {
      const response = await saveProfile(data);

      if (!response.ok) {
        toast({
          title: "Failed to save profile",
          variant: "destructive",
        });
        console.error("error", response);
      }

      toast({
        title: "Profile Saved!",
        variant: "success",
      });
    } catch (error) {
      toast({
        title: "Failed to save profile",
        variant: "destructive",
      });
      console.error("error", error);
    } finally {
      setLoading(false);
    }
  };

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
        <FormField
          control={form.control}
          name="github_handle"
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
          name="twitter_handle"
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
          name="linked_in_handle"
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
        <Button type="submit">
          {loading && <Icons.spinner className="mr-2 size-5 animate-spin" />}
          Update profile
        </Button>
      </form>
    </Form>
  );
}
