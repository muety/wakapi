import { getData } from "@/actions/api";
import { LeaderBoardTable } from "@/components/leaderboard-table";
import { LeaderboardApiResponse } from "@/lib/types";

export default async function Leaderboards({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const queryParams = new URLSearchParams(searchParams);
  const url = `/v1/leaders?${queryParams.toString()}`;
  const durationData = await getData<LeaderboardApiResponse>(url, {
    skipAuth: true,
  });

  if (!durationData.success) {
    return <div>There was an error fetching leaderboard data...</div>;
  }

  return (
    <LeaderBoardTable
      data={durationData.data}
      title="Top Coders"
      titleClass="text-center mb-8 text-6xl"
    />
  );
}
