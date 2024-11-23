import { redirectIfLoggedIn } from "@/actions";
import { Hero } from "./hero";
import { Rubik } from "next/font/google";
const rubik = Rubik({ subsets: ["latin"] });

export default async function IndexPage() {
  await redirectIfLoggedIn();
  return (
    <div className={rubik.className}>
      <Hero />
    </div>
  );
}
