import { FadeOnView } from "./fade-on-view";
import InstallationInstructions from "./installation-instructions";

export default function HowItWorks() {
  return (
    <FadeOnView>
      <section className="py-12 md:py-16 lg:py-20">
        <div className="container space-y-5">
          <h2 className="text-3xl font-bold tracking-tight mb-12 text-center">
            How it Works
          </h2>
          <InstallationInstructions />
        </div>
      </section>
    </FadeOnView>
  );
}
