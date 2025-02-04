import { format, subDays } from "date-fns";
import Image from "next/image";

import { fetchData, getSession } from "@/actions";
import { ProjectBreadCrumb } from "@/components/bread-crumbs";
import { ActivityCategoriesChart } from "@/components/charts/ActivityCategoriesChart";
import { DailyCodingSummaryLineChart } from "@/components/charts/DailyCodingSummaryLineChart";
import { WPieChart } from "@/components/charts/WPieChart";
import { DashboardPeriodSelector } from "@/components/dashboard-period-selector";
import { ProjectFiles } from "@/components/ProjectFiles";
import { SummariesApiResponse } from "@/lib/types";
import { makePieChartDataFromRawApiResponse } from "@/lib/utils";

const { NEXT_PUBLIC_API_URL } = process.env;

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
    <div className="my-6">
      {durationData && !(durationData instanceof Error) && (
        <main>
          <div className="flex items-center justify-between align-middle">
            <ProjectBreadCrumb projectId={params.id} />
            <div>
              {/* <img
                className="with-url-src"
                src={`${NEXT_PUBLIC_API_URL}/api/badge/${session.user.id}/project:${params.id}/interval:all_time?label=total&token=${session.token}`}
                alt="Badge"
              /> */}
              <Image
                className="with-url-src"
                src={`${NEXT_PUBLIC_API_URL}/api/badge/${session.user.id}/project:${params.id}/interval:all_time?label=total&token=${session.token}`}
                alt="Badge"
                width={150}
                height={20}
                unoptimized
              />
            </div>
          </div>
          <div className="m-0 mb-5 mt-2 text-lg">
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
          <div className="mt-12 flex justify-center gap-5">
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
