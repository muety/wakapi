import { fetchData } from "@/actions";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { UserPreferences } from "@/components/user-preferences";
import { UserProfile } from "@/lib/types";

export default async function Page() {
  const user = await fetchData<UserProfile>("/profile", true);
  return (
    <div className="grid gap-6">
      <Card x-chunk="dashboard-04-chunk-1">
        <CardHeader>
          <CardTitle className="text-2xl">Preferences</CardTitle>
          <CardDescription>Edit your account preferences.</CardDescription>
        </CardHeader>
        <CardContent>
          <UserPreferences user={user!} />
        </CardContent>
      </Card>
    </div>
  );
}
