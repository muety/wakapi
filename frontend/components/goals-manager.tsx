"use client";

import { useEffect, useState } from "react";

import { GoalData, Project } from "@/lib/types";

import { AddGoalDialogV2 } from "./add-goal-v2";
import { Goal } from "./goal";

interface IProps {
  goals: GoalData[];
  projects: Project[];
}

export default function GoalsManager({
  goals: initialGoals,
  projects,
}: IProps) {
  const [goals, setGoals] = useState(initialGoals);
  const [removingGoalId, setRemovingGoalId] = useState<string | null>(null);

  const onDeleteGoal = (deletedGoal: GoalData) => {
    setRemovingGoalId(deletedGoal.id);
    setTimeout(() => {
      setGoals(goals.filter((goal) => goal.id !== deletedGoal.id));
      setRemovingGoalId(null);
    }, 500);
  };

  const onAddGoal = (newGoal: GoalData) => {
    setGoals([newGoal, ...goals]);
  };

  useEffect(() => {
    goals.forEach((goal) => {
      const element = document.getElementById(goal.id);
      if (element) {
        element.classList.add("slide-in");
      }
    });
  }, [goals]);

  return (
    <div className="mx-2 my-6">
      <div className="mb-5 flex items-center justify-between">
        <h1 className="text-4xl">Work Goals</h1>
        <AddGoalDialogV2 onAddGoal={onAddGoal} projects={projects} />
      </div>
      <div className="flex flex-col gap-6">
        {goals.map((goal, index) => (
          <div
            key={goal?.id || index}
            id={goal?.id || `goal-${index}`}
            className={removingGoalId === goal?.id ? "slide-out" : ""}
          >
            <Goal data={goal} onDeleteGoal={onDeleteGoal} />
          </div>
        ))}
      </div>
    </div>
  );
}
