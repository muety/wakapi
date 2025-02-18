import * as z from "zod";

export const userNameSchema = z.object({
  email: z.string().min(3).max(32),
  password: z.string().min(3).max(32),
});

export const otpLoginSchema = z.object({
  email: z.string().min(3).max(32),
  code_challenge: z.string().min(3),
  challenge_method: z.string().min(3),
});

export const otpLoginValidateSchema = z.object({
  email: z.string().min(3).max(32),
  code_verifier: z.string().min(3),
  otp: z.string().min(3),
});

export const resetPasswordSchema = z.object({
  password: z.string().min(3).max(32),
  confirm_password: z.string().min(3).max(32),
});

export const forgotPasswordSchema = z.object({
  email: z.string().min(3).max(32),
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
