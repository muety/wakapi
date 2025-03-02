import Image from "next/image";

import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";
import useSession from "@/lib/session/use-session";

import { Project } from "./projects-table";

export function ProjectTimeImage({ project }: { project: Project }) {
  const { session } = useSession();

  if (!session) {
    null
  }

  const src = `${NEXT_PUBLIC_API_URL}/api/badge/${session.user.id}/project:${project.id}/interval:all_time?label=total&token=${session.token}`
  return (
    <Image
      className="with-url-src"
      src={src}
      alt="Badge"
      width={120}
      height={15}
      unoptimized
    />
  );
}
