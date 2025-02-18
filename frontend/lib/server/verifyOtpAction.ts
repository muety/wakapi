// // import { actionClient } from "@/lib/actionClient";
// import { z } from "zod";
// import { actionClient } from "./actionClient";

// export const verifyOtpAction = actionClient.(
//   z.object({
//     email: z.string().email(),
//     otp: z.string(),
//     code_verifier: z.string(),
//   }),
//   async ({
//     email,
//     otp,
//     code_verifier,
//   }: {
//     email: string;
//     otp: string;
//     code_verifier: string;
//   }) => {
//     // Perform OTP verification logic (e.g., database lookup)
//     return {
//       success: true,
//       message: "OTP verified",
//       email,
//       otp,
//       code_verifier,
//     };
//   }
// );
