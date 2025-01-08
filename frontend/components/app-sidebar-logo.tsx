import Image from "next/image";

export function AppSidebarLogo() {
  return (
    <div className="sidebar-logo py-5 pl-1">
      <Image
        src={"/white-logo.png"}
        alt="Logo"
        width={150}
        height={34}
        className="expanded"
      />
      <Image
        src={"/white-icon.svg"}
        alt="Logo"
        width={43}
        height={31}
        className="collapsed"
      />
    </div>
  );
}
