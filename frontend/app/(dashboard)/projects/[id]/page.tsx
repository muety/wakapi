import { ProjectFiles } from "@/components/ProjectFiles";
import { WPieChart } from "@/components/charts/WPieChart";
import { DashboardPeriodSelector } from "@/components/dashboard-period-selector";
import { ActivityCategoriesChart } from "@/components/charts/ActivityCategoriesChart";
import { DailyCodingSummaryLineChart } from "@/components/charts/DailyCodingSummaryLineChart";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { format, subDays } from "date-fns";
import { fetchData, getSession } from "@/actions";
import { SummariesApiResponse } from "@/lib/types";
import { makePieChartDataFromRawApiResponse } from "@/lib/utils";

const { API_URL } = process.env;

export function ProjectBreadCrumb({ projectId }: { projectId: string }) {
  return (
    <Breadcrumb className="text-2xl mb-4 ml-0 pl-0 m-0">
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink
            className="link text-xl hover:text-purple underline"
            href="/projects"
          >
            Projects
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbLink
            className="link text-xl hover:text-purple"
            href={`/projects/${projectId}`}
          >
            {projectId}
          </BreadcrumbLink>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  );
}

interface iProps {
  searchParams: Record<string, any>;
  params: { id: string };
}

export default async function ProjectDetailPage({
  params,
  searchParams,
}: iProps) {
  const session = await getSession();

  searchParams["project"] = params.id;
  const projectUrl = `/projects/${params.id}`;

  const {
    start = format(subDays(new Date(), 7), "yyyy-MM-dd"),
    end = format(new Date(), "yyyy-MM-dd"),
  } = searchParams;

  const url = `/compat/wakatime/v1/users/current/summaries?${new URLSearchParams(
    { start, end, project: params.id }
  )}`;
  const durationData = await fetchData<SummariesApiResponse>(url);
  return (
    <div className="p-4 white-well px-6 my-6 mx-2">
      {durationData && !(durationData instanceof Error) && (
        <main>
          <div className="flex justify-between align-middle items-center">
            <ProjectBreadCrumb projectId={params.id} />
            <div>
              <img
                className="with-url-src"
                src={`${API_URL}/api/badge/${session.user.id}/project:${params.id}/interval:all_time?label=total&token=${session.token}`}
                alt="Badge"
              />
            </div>
          </div>
          <div className="m-0 text-lg mb-5 mt-2">
            <b>{durationData.cumulative_total.text}</b>{" "}
            <span>over the last</span>{" "}
            <DashboardPeriodSelector
              searchParams={searchParams}
              baseUrl={projectUrl}
            />{" "}
            <span>in {params.id}</span>
          </div>
          <section className="charts-grid">
            <div className="min-h-52">
              <DailyCodingSummaryLineChart data={durationData.data} />
            </div>
            <div className="min-h-52">
              <ActivityCategoriesChart data={durationData.data} />
            </div>
          </section>
          <section className="charts-grid">
            <div>
              <WPieChart
                title="Languages"
                data={makePieChartDataFromRawApiResponse(
                  durationData.data,
                  "languages"
                )}
                colorNamespace="languages"
                durationSubtitle="Languages used over the "
              />
            </div>
            <div>
              <WPieChart
                title="Editors"
                data={makePieChartDataFromRawApiResponse(
                  durationData.data,
                  "languages"
                )}
                colorNamespace="languages"
                durationSubtitle="Languages used over the "
              />
            </div>
          </section>
          <div className="flex justify-center gap-5 mt-12">
            <div className="flex justify-between gap-40">
              <ProjectFiles
                data={durationData.data}
                field="entities"
                title="Files"
                showCopy={true}
              />
              <ProjectFiles
                data={durationData.data}
                field="branches"
                title="Branches"
              />
            </div>
          </div>
        </main>
      )}
    </div>
  );
}
