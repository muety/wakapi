import * as z from "zod";

export const userNameSchema = z.object({
  email: z.string().min(3).max(32),
  password: z.string().min(3).max(32),
});

export const userSignupSchema = z
  .object({
    email: z.string().min(3).max(32),
    password: z.string().min(3).max(32),
    password_repeat: z.string().min(3).max(32),
  })
  .refine(
    ({ password, password_repeat }) => {
      return password === password_repeat;
    },
    {
      message: "Passwords do not match",
      path: ["password_repeat"],
    }
  );
