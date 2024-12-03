import { Rubik } from "next/font/google";

import { redirectIfLoggedIn } from "@/actions";

import { Hero } from "./hero";
import { Metadata } from "next";
const rubik = Rubik({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Homepage",
  description: "Wakana homepage",
};

export default async function IndexPage() {
  await redirectIfLoggedIn();
  return (
    <div className={rubik.className}>
      <Hero />
    </div>
  );
}
