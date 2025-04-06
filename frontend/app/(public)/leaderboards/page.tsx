import { fetchData } from "@/actions";
import { LeaderBoardTable } from "@/components/leaderboard-table";
import { LeaderboardApiResponse } from "@/lib/types";

export default async function Leaderboards({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const queryParams = new URLSearchParams(searchParams);
  const url = `/v1/leaders?${queryParams.toString()}`;
  const durationData = await fetchData<LeaderboardApiResponse>(url, false);

  if (!durationData) {
    return <div>There was an error fetching leaderboard data...</div>;
  }

  return (
    <LeaderBoardTable
      data={durationData}
      title="Top Coders"
      titleClass="text-center mb-8 text-6xl"
    />
  );
}
