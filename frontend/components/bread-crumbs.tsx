import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export function ProjectBreadCrumb({ projectId }: { projectId: string }) {
  return (
    <Breadcrumb className="m-0 mb-4 pl-0 text-2xl">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink
            className="link hover:text-purple text-xl underline"
            href="/projects"
          >
            Projects
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbLink
            className="link hover:text-purple text-xl"
            href={`/projects/${projectId}`}
          >
            {projectId}
          </BreadcrumbLink>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  );
}
