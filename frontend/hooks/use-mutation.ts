import { useState } from "react";
import { toast } from "@/components/ui/use-toast";
import { postData, updateData, deleteData } from "@/actions/api";

type MutationMethod = "post" | "put" | "delete";

interface UseMutationOptions<TData, TVariables> {
  onSuccess?: (data: TData) => void;
  onError?: (error: Error) => void;
  successMessage?: string | ((data: TData) => string);
  errorMessage?: string | ((error: Error) => string);
  skipToast?: boolean;
}

export function useMutation<TData = any, TVariables = any>(
  path: string,
  method: MutationMethod = "post",
  options: UseMutationOptions<TData, TVariables> = {}
) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<TData | null>(null);

  const mutate = async (variables?: TVariables): Promise<TData | null> => {
    setIsLoading(true);
    setError(null);

    try {
      let response;

      // Be explicit about which function signature we're using
      if (method === "post") {
        response = await postData<TData, TVariables>(
          path,
          variables as TVariables
        );
        console.log("POST RESPONSE", response);
      } else if (method === "put") {
        response = await updateData<TData, TVariables>(
          path,
          variables as TVariables
        );
      } else {
        // For delete, we might not have a body
        response = await deleteData<TData>(path);
      }

      if (!response.success) {
        throw new Error(response.error?.message || "Operation failed");
      }

      // Handle success
      const responseData = response.data;
      if (responseData) {
        setData(responseData);
      }
      if (options.onSuccess && responseData) {
        options.onSuccess(responseData);
      }

      // Show success toast
      if (!options.skipToast) {
        const message =
          typeof options.successMessage === "function"
            ? options.successMessage(response.data)
            : options.successMessage || "Operation completed successfully";

        toast({
          title: "Success",
          description: message,
          variant: "success",
        });
      }

      return response.data;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);

      // Handle error
      if (options.onError) {
        options.onError(error);
      }

      // Show error toast
      if (!options.skipToast) {
        const message =
          typeof options.errorMessage === "function"
            ? options.errorMessage(error)
            : options.errorMessage || error.message || "Operation failed";

        toast({
          title: "Error",
          description: message,
          variant: "destructive",
        });
      }

      return null;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    mutate,
    isLoading,
    error,
    data,
    reset: () => {
      setData(null);
      setError(null);
    },
  };
}
