import { toast } from "@/components/ui/use-toast";

export const copyApiKeyToClickBoard = async (apiKey: string) => {
  try {
    await navigator.clipboard.writeText(apiKey);
    toast({
      title: "Api key copied to clipboard",
      variant: "success",
    });
  } catch (error) {
    toast({
      title: (error as Error).message || "Api key failed",
      description:
        "You're likely using an old browser that doesn't support this feature.",
      variant: "destructive",
    });
  }
};
