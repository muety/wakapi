import { ApiClient } from "@/actions/api";
import { LeaderBoardTable } from "@/components/leaderboard-table";
import { LeaderboardApiResponse } from "@/lib/types";

export default async function Leaderboards({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const queryParams = new URLSearchParams(searchParams);
  const url = `/v1/leaders?${queryParams.toString()}`;
  const durationData = await ApiClient.GET<LeaderboardApiResponse>(url, {
    skipAuth: true,
  });

  if (!durationData.success) {
    return <div>There was an error fetching leaderboard data...</div>;
  }

  console.log("durationData", durationData);

  return (
    <div className="mt-5 min-h-screen">
      <LeaderBoardTable data={durationData.data} title="Public" />
    </div>
  );
}
