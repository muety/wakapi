import { fetchData } from "@/actions";
import { LeaderBoardTable } from "@/components/leaderboard-table";
import { LeaderboardApiResponse } from "@/lib/types";

export default async function Leaderboards({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const queryParams = new URLSearchParams(searchParams);
  const url = `compat/wakatime/v1/leaders?${queryParams.toString()}`;
  console.log("url", url);
  const durationData = await fetchData<LeaderboardApiResponse>(url, false);

  if (!durationData) {
    return <div>There was an error fetching leaderboard data...</div>;
  }

  return (
    <div className="mt-5 min-h-screen">
      <LeaderBoardTable data={durationData} title="Public" />
    </div>
  );
}
