import { LucideGithub, LucideTwitter } from "lucide-react";

export function SimpleFooter() {
  return (
    <div className="main-bg max-h-24 md:-mx-8 md:px-8">
      <hr className="shadow-sm" />
      <div className="simple-footer flex flex-col items-center justify-between px-2 pb-5 pt-10 align-middle md:flex-row">
        <div>
          <ul className="flex flex-col gap-4 md:flex-row">
            <li className="credit">© {new Date().getFullYear()} Wakana</li>
            <li>
              <a href="/terms">Terms</a>
            </li>
            <li>
              <a href="/privacy">Privacy</a>
            </li>
            <li>
              <a href="/about">About</a>
            </li>
            <li>
              <a href="/blog">Blog</a>
            </li>
          </ul>
        </div>
        <div className="hidden justify-start md:flex md:justify-center">
          <ul className="flex flex-col gap-1 md:flex-row">
            <li>
              <a href="https://github.com/jemiluv8" rel="noopener noreferrer">
                {/* <i className="fa fa-github-square"></i> */}
                <LucideGithub className="rounded-md border-2 border-black" />
              </a>
            </li>
            <li>
              <a
                href="https://twitter.com/intent/user?screen_name=WakaTimer"
                rel="noopener noreferrer"
              >
                {/* <i className="fa fa-twitter-square"></i> */}
                <LucideTwitter className="rounded-md border-2 border-blue-500 text-blue-500" />
              </a>
            </li>
          </ul>
        </div>
        <div>
          <ul className="flex flex-col gap-4 align-middle md:flex-row">
            <li>
              <a href="/leaders">Leaderboard</a>
            </li>
            <li>
              <a href="/faqs">FAQs</a>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
