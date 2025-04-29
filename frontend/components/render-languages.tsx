"use client";

import { truncate } from "lodash";
import Link from "next/link";
import { usePathname } from "next/navigation";

export function RenderLanguages({ languages }: { languages: string[] }) {
  const truncated = truncate(languages.join(", "), { length: 75 });
  const truncatedArray = truncated.split(", ");
  const last_item = truncatedArray.pop();
  const currentPath = usePathname();

  const items = truncatedArray.map((item, index) => (
    <a
      key={index}
      href={`${currentPath}?language=${item}`}
      className="gap-2 text-white hover:underline"
    >
      {item},
    </a>
  ));
  const totalCharacters = truncatedArray.reduce(
    (acc, item) => acc + item.length + 1,
    0
  );
  const remainingCharacters = 300 - totalCharacters;
  const lastItemText = truncate(last_item, { length: remainingCharacters });
  return (
    <div className="flex" style={{ gap: "1px" }} title={languages.join(", ")}>
      {items}
      {last_item && (
        <Link
          href={`leaderboards?language=${last_item}`}
          className="text-white hover:underline"
        >
          {lastItemText}
        </Link>
      )}
    </div>
  );
}
