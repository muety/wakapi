"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

interface Props {
  title: string;
  href: string;
}

export function SettingItem(props: Props) {
  // get current route
  const classList = ["settings", "text-primary"];
  //   const pathname = window ? window.location.href : "";
  const pathName = usePathname();

  if (pathName === props.href) {
    classList.push("active");
  }

  return (
    <div className={classList.join(" ")}>
      <Link href={props.href} style={{ margin: 0, padding: 0 }}>
        {props.title}
      </Link>
    </div>
  );
}
