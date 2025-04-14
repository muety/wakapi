import * as z from "zod";

export const profileFormSchema = z.object({
  name: z
    .string()
    .min(2, {
      message: "Name must be at least 2 characters.",
    })
    .optional(),
  username: z
    .string()
    .min(2, {
      message: "Username must be at least 2 characters.",
    })
    .optional(),
  bio: z
    .string()
    .max(160, {
      message: "Bio must not be longer than 160 characters.",
    })
    .optional(),
  github_handle: z
    .string()
    .regex(/^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$/i, {
      message: "Please enter a valid GitHub username.",
    })
    .optional(),
  twitter_handle: z
    .string()
    .regex(/^[a-zA-Z0-9_]{1,15}$/, {
      message: "Please enter a valid Twitter username.",
    })
    .optional(),
  linked_in_handle: z
    .string()
    .regex(/^[a-zA-Z0-9-]{3,100}$/, {
      message: "Please enter a valid LinkedIn username.",
    })
    .optional(),
});

export type ProfileFormValues = z.infer<typeof profileFormSchema>;
