import { fetchData } from "@/actions";
import { ProfileForm } from "@/components/profile-form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { UserProfile } from "@/lib/types";

export default async function ProfilePage() {
  const user = await fetchData<UserProfile>("/profile", true);

  return (
    <Card x-chunk="dashboard-04-chunk-1">
      <CardHeader>
        <CardTitle className="text-2xl">Profile</CardTitle>
        <CardDescription>
          This is your public profile. Only share want you want made public.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ProfileForm user={user!} />
      </CardContent>
    </Card>
  );
}
