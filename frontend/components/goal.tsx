"use client";

import { LucideSave, LucideTrash2, LucideUndo } from "lucide-react";
import React from "react";
import useSWRMutation from "swr/mutation";

import { toast } from "@/components/ui/use-toast";
import { mutateData } from "@/hooks/api-utils";
import { GoalData } from "@/lib/types";

import { GoalChart } from "./charts/GoalChart";
import { Icons } from "./icons";
import { Button } from "./ui/button";
import { Card, CardTitle } from "./ui/card";
import { Confirm } from "./ui/confirm";

interface iProps {
  data: GoalData;
  token: string;
  onDeleteGoal: (goal: GoalData) => void;
}

export function Goal({ data, token, onDeleteGoal }: iProps) {
  const originalText = data.custom_title || data.title;
  const [title, setTitle] = React.useState(originalText);

  const resourceUrl = `/v1/users/current/goals/${data.id}`;

  const { trigger, isMutating: loading } = useSWRMutation(
    resourceUrl,
    mutateData
  );

  const { trigger: updateTrigger, isMutating } = useSWRMutation(
    resourceUrl,
    mutateData
  );

  const updateGoal = async () => {
    console.log("Updating goal...");
    if (title !== originalText) {
      try {
        const response = await updateTrigger({
          method: "PUT",
          body: { title },
          token,
        });
        console.log("response", response);
        toast({
          title: "Goal updated",
          variant: "success",
        });
      } catch (error) {
        toast({
          title: "Failed to update goal",
          variant: "destructive",
          description: (error as Error).message,
        });
      }
    }
  };

  const deleteGoal = async () => {
    try {
      await trigger({
        method: "DELETE",
        token,
      });
      toast({
        title: "Deleted",
        description: `Goal with title: ${title} - deleted`,
        variant: "success",
      });
      onDeleteGoal(data);
    } catch (error) {
      toast({
        title: "Failed to delete goal",
        variant: "destructive",
        description: (error as Error).message,
      });
    }
  };

  return (
    <Card>
      <div className="goal-container">
        <CardTitle className="ml-5">
          <div
            className="flex h-12 items-center justify-end"
            style={{ justifyItems: "stretch" }}
          >
            <div className="goal-title-viewer">
              <input
                value={title}
                onChange={(event) => setTitle(event.target.value)}
              ></input>
            </div>
            <div className="ml-2 flex items-center gap-1 bg-muted">
              {title !== originalText && (
                <div className="flex">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => {
                      console.log("Saving goal...");
                      updateGoal();
                    }}
                  >
                    {!(loading || isMutating) ? (
                      <LucideSave className="size-4 cursor-pointer text-gray-400" />
                    ) : (
                      <Icons.spinner className="mr-2 size-4 animate-spin" />
                    )}
                  </Button>
                  <Button
                    onClick={() => setTitle(originalText)}
                    variant="ghost"
                    size="icon"
                  >
                    <LucideUndo className="size-4 cursor-pointer text-gray-400" />
                  </Button>
                </div>
              )}
              {title === originalText && (
                <Confirm
                  title="Delete Goal"
                  description={`Delete goal: ${title}`}
                  onConfirm={() => deleteGoal()}
                  loading={loading}
                >
                  <Button
                    variant="ghost"
                    size="icon"
                    className="goal-item-icon"
                  >
                    <LucideTrash2 className="size-4 cursor-pointer text-red-400" />
                  </Button>
                </Confirm>
              )}
            </div>
          </div>
        </CardTitle>
        <GoalChart
          data={data.chart_data}
          direction={data.target_direction || "more"}
        />
      </div>
    </Card>
  );
}
