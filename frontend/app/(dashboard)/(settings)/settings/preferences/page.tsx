import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";

export default async function Page() {
  return (
    <div className="grid gap-6">
      <Card x-chunk="dashboard-04-chunk-1">
        <CardHeader>
          <CardTitle>Preferences</CardTitle>
          <CardDescription>Edit your account preferences.</CardDescription>
        </CardHeader>
        <CardContent>
          <p>Missing content</p>
        </CardContent>
      </Card>
    </div>
  );
}
