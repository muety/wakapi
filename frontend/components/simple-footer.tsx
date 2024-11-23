import { LucideGithub, LucideTwitter } from "lucide-react";

export function SimpleFooter() {
  return (
    <div className="main-bg md:px-8 max-h-24 md:-mx-8">
      <hr className="shadow-sm" />
      <div className="flex flex-col md:flex-row justify-between align-middle px-2 pt-10 pb-5 items-center simple-footer">
        <div>
          <ul className="flex flex-col md:flex-row gap-4">
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
        <div className="md:flex justify-start md:justify-center hidden">
          <ul className="flex gap-1 flex-col md:flex-row">
            <li>
              <a href="https://github.com/jemiluv8" rel="noopener noreferrer">
                {/* <i className="fa fa-github-square"></i> */}
                <LucideGithub className="border-black border-2 rounded-md" />
              </a>
            </li>
            <li>
              <a
                href="https://twitter.com/intent/user?screen_name=WakaTimer"
                rel="noopener noreferrer"
              >
                {/* <i className="fa fa-twitter-square"></i> */}
                <LucideTwitter className="border-blue-500 border-2 rounded-md text-blue-500" />
              </a>
            </li>
          </ul>
        </div>
        <div>
          <ul className="flex gap-4 flex-col md:flex-row align-middle">
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
