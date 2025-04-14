import Image from "next/image";
import Link from "next/link";

export default async function Page({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="container flex h-screen min-h-screen w-screen flex-col items-center justify-center gap-4">
      <Link href="/">
        <Image
          autoCorrect="on"
          src={"/white-logo.svg"}
          alt="Logo"
          width={178}
          height={40.84}
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
