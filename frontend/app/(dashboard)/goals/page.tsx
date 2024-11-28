import { fetchData, getSession } from "@/actions";
import GoalsManager from "@/components/goals-manager";
import { GoalData, Project } from "@/lib/types";

export default async function Goals() {
  const session = await getSession();

  const goalData = await fetchData<{ data: GoalData[] }>(
    "compat/wakatime/v1/users/current/goals"
  );
  if (!goalData) {
    return <p>Error fetching goals</p>;
  }

  const projects = await fetchData<{ data: Project[] }>(
    "compat/wakatime/v1/users/current/projects"
  );

  const goals = goalData.data;

  return (
    <div className="panel panel-default mx-2 my-6 min-h-screen min-w-full p-2 px-6">
      <GoalsManager
        goals={goals}
        token={session.token}
        projects={projects?.data || []}
      />
    </div>
  );
}
