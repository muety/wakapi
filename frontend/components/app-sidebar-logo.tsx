import Image from "next/image";

export function AppSidebarLogo() {
  return (
    <div className="sidebar-logo pl-1 py-5">
      <Image
        src={"/logo/white-logo.png"}
        alt="Logo"
        width={150}
        height={34}
        className="expanded"
      />
      <Image
        src={"/logo/white-icon.png"}
        alt="Logo"
        width={43}
        height={31}
        className="collapsed"
      />
    </div>
  );
}
