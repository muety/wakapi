import { Metadata } from "next";
import { Rubik } from "next/font/google";

import { redirectIfLoggedIn } from "@/actions";
import FeatureSection from "@/components/features-hero";
import HowItWorks from "@/components/how-it-works";

import { Hero } from "./hero";
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
      <FeatureSection />
      <HowItWorks />
    </div>
  );
}
