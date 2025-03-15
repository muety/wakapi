import { Github } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

export default function PublicFooter() {
  return (
    <footer
      data-state="false"
      className="flex w-full border-t-[0.5px] border-border-1 flex-col items-center pt-[5rem] pb-10 md:pb-32 footer data-[state=false]:mt-8 "
    >
      <div className="flex w-full lg:max-w-[73.25rem] justify-between px-4 md:px-0">
        <div>
          <div className="flex gap-2">
            <Image width={200} height={100} src="/white-logo.png" alt="Logo" />
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 md:gap-16">
          <div>
            <h3 className="text-base font-medium mb-4">Shortcuts</h3>
            <ul className="space-y-3">
              <li>
                <Link
                  href="/faqs"
                  className="text-gray-400 hover:text-white transition-colors"
                >
                  FAQS
                </Link>
              </li>
              <li>
                <Link
                  href="/plugins"
                  className="text-gray-400 hover:text-white transition-colors"
                >
                  Plugin
                </Link>
              </li>
              <li>
                <Link
                  href="/pricing"
                  className="text-gray-400 hover:text-white transition-colors"
                >
                  Pricing
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-base font-medium mb-4">Company</h3>
            <ul className="space-y-3">
              <li>
                <Link
                  href="https://github.com/jemiluv8/wakana"
                  className="text-gray-400 hover:text-white transition-colors flex items-center gap-2"
                >
                  <Github size={18} />
                  <span>Github</span>
                </Link>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </footer>
  );
}
