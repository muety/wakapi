import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { UserPreferences } from "@/components/user-preferences";

export default async function Page() {
  return (
    <div className="grid gap-6">
      <Card x-chunk="dashboard-04-chunk-1">
        <CardHeader>
          <CardTitle className="text-2xl">Preferences</CardTitle>
          <CardDescription>Edit your account preferences.</CardDescription>
        </CardHeader>
        <CardContent>
          <UserPreferences />
        </CardContent>
      </Card>
    </div>
  );
}
