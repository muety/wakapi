import { ApiKeyCopier } from "@/components/copy-api-key";
import { ProfileForm } from "@/components/profile-form";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";

export default function ProfilePage() {
  return (
    <Card x-chunk="dashboard-04-chunk-1">
      <CardHeader>
        <CardTitle className="text-2xl">Profile</CardTitle>
        <CardDescription>
          This is your public profile. Only share want you want made public.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ProfileForm />
      </CardContent>
    </Card>
  );
}
