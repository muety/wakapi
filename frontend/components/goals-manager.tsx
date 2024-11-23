"use client";

import React, { useState, useEffect } from "react";
import { Goal } from "./goal";
import { GoalData, Project } from "@/lib/types";
import { AddGoalDialogV2 } from "./add-goal-v2";

interface IProps {
  token: string;
  goals: GoalData[];
  projects: Project[];
}

export default function GoalsManager({
  goals: initialGoals,
  token,
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
      <div className="flex justify-between items-center mb-5">
        <h1 className="text-4xl">Work Goals</h1>
        <AddGoalDialogV2
          onAddGoal={onAddGoal}
          projects={projects}
          token={token}
        />
      </div>
      <div className="flex flex-col gap-6">
        {goals.map((goal) => (
          <div
            key={goal.id}
            id={goal.id}
            className={removingGoalId === goal.id ? "slide-out" : ""}
          >
            <Goal data={goal} token={token} onDeleteGoal={onDeleteGoal} />
          </div>
        ))}
      </div>
    </div>
  );
}
