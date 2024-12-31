interface AuthLayoutProps {
  children: React.ReactNode;
}

import Image from "next/image";
import Link from "next/link";

export default async function Page({ children }: AuthLayoutProps) {
  return (
    <div className="container flex h-screen min-h-screen w-screen flex-col items-center justify-center gap-4">
      <Link href="/">
        <Image
          src={"/white-logo.png"}
          alt="Logo"
          width={128}
          height={70}
          className="logo-icon-white"
        />
      </Link>
      <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
        <div className="flex flex-col justify-center justify-items-center space-y-2 text-center align-middle">
          {children}
        </div>
      </div>
    </div>
  );
}
