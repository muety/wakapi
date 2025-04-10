"use client";

import React from "react";
import styles from "./add-goal.module.css";
import WMultiSelect from "./w-multi-select";

import { LucidePlus } from "lucide-react";
import { useReducer } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "@/components/ui/use-toast";
import {
  CATEGORY_OPTIONS,
  EDITOR_OPTIONS,
  LANGUAGE_OPTIONS,
} from "@/lib/constants";
import { GoalData, Project } from "@/lib/types";
import { cn } from "@/lib/utils";

import { ClickToSelect } from "./click-to-select";
import { Icons } from "./icons";
import { SimpleSelect } from "./simple-select";
import { Input } from "./ui/input";
import { useMutation } from "@/hooks/use-mutation";

enum GoalActionType {
  RESET = "RESET",
  SET_CODE_MORE = "SET_CODE_MORE",
  SET_LOADING = "SET_LOADING",
  SET_DIALOG_OPEN = "SET_DIALOG_OPEN",
  SET_TARGET_DURATION = "SET_TARGET_DURATION",
  SET_SELECTED_GOAL_OPTION = "SET_SELECTED_GOAL_OPTION",
  SET_TARGET_DURATION_TYPE = "SET_TARGET_DURATION_TYPE",
  SET_TARGET_DURATION_PERIOD = "SET_TARGET_DURATION_PERIOD",
  SET_SELECTED_LANGUAGES = "SET_SELECTED_LANGUAGES",
  SET_SELECTED_EDITORS = "SET_SELECTED_EDITORS",
  SET_SELECTED_CATEGORIES = "SET_SELECTED_CATEGORIES",
  SET_SELECTED_PROJECTS = "SET_SELECTED_PROJECTS",
  SET_HAS_COMPLETED_SECOND_STEP = "SET_HAS_COMPLETED_SECOND_STEP",
}

// Define the GoalState type
type GoalState = {
  codeMore: string;
  loading: boolean;
  dialogOpen: boolean;
  targetDuration: string;
  selectedGoalOption: string;
  targetDurationType: string;
  targetDurationPeriod: string;
  selectedLanguages: string[];
  selectedEditors: string[];
  selectedCategories: string[];
  selectedProjects: string[];
  hasCompletedSecondStep: boolean;
};

const initialState: GoalState = {
  codeMore: "",
  loading: false,
  dialogOpen: false,
  targetDuration: "",
  selectedGoalOption: "",
  targetDurationType: "hrs",
  targetDurationPeriod: "day",
  selectedLanguages: [],
  selectedEditors: [],
  selectedCategories: [],
  selectedProjects: [],
  hasCompletedSecondStep: false,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
enum GoalOption {
  overall = "overall",
  language = "language",
  editor = "editor",
  category = "category",
  project = "project",
}

type GoalOptions = keyof typeof GoalOption;
export type ConfigurableGoalOptions = Exclude<
  keyof typeof GoalOption,
  "overall"
>;

type GoalOptionItem = {
  label: string;
  value: GoalOptions;
  id: string;
};

function reducer(
  state: GoalState,
  action: { type: GoalActionType; payload: any }
): GoalState {
  switch (action.type) {
    case GoalActionType.RESET:
      return initialState;
    case GoalActionType.SET_CODE_MORE:
      return { ...state, codeMore: action.payload };
    case GoalActionType.SET_LOADING:
      return { ...state, loading: action.payload };
    case GoalActionType.SET_DIALOG_OPEN:
      return { ...state, dialogOpen: action.payload };
    case GoalActionType.SET_TARGET_DURATION:
      return { ...state, targetDuration: action.payload };
    case GoalActionType.SET_SELECTED_GOAL_OPTION:
      return { ...state, selectedGoalOption: action.payload };
    case GoalActionType.SET_TARGET_DURATION_TYPE:
      return { ...state, targetDurationType: action.payload };
    case GoalActionType.SET_TARGET_DURATION_PERIOD:
      return { ...state, targetDurationPeriod: action.payload };
    case GoalActionType.SET_SELECTED_LANGUAGES:
      return { ...state, selectedLanguages: action.payload };
    case GoalActionType.SET_SELECTED_EDITORS:
      return { ...state, selectedEditors: action.payload };
    case GoalActionType.SET_SELECTED_CATEGORIES:
      return { ...state, selectedCategories: action.payload };
    case GoalActionType.SET_SELECTED_PROJECTS:
      return { ...state, selectedProjects: action.payload };
    case GoalActionType.SET_HAS_COMPLETED_SECOND_STEP:
      return { ...state, hasCompletedSecondStep: action.payload };
    default:
      return state;
  }
}

export function AddGoalDialogV2({
  onAddGoal,
  projects,
}: {
  onAddGoal: (goal: any) => void;
  projects: Project[];
}) {
  const [state, dispatch] = useReducer(reducer, initialState);

  const {
    codeMore,
    dialogOpen,
    targetDuration,
    selectedGoalOption,
    targetDurationType,
    targetDurationPeriod,
    selectedLanguages,
    selectedEditors,
    selectedCategories,
    selectedProjects,
  } = state;

  const getDurationSeconds = React.useMemo(() => {
    const duration = parseInt(targetDuration);
    if (targetDurationType === "hrs") {
      return duration * 3600;
    }
    if (targetDurationType === "mins") {
      return duration * 60;
    }
    return duration;
  }, [targetDuration, targetDurationType]);

  const { mutate: createGoalMethod, isLoading: loading } = useMutation<
    { data: GoalData },
    any
  >(`/v1/users/current/goals`, "post", {
    successMessage: "Goal created successfully",
    onSuccess: (newGoal) => {
      console.log("Goal created successfully", newGoal);
      onAddGoal(newGoal.data);
      dispatch({ type: GoalActionType.RESET, payload: null });
    },
    // errorMessage: "Failed to create goal",
  });

  const createGoal = async () => {
    if (!targetDuration || +targetDuration <= 0 || !canSpecifyDuration) {
      toast({
        title: "Invalid input",
        description: "Please check your goal configuration",
        variant: "destructive",
      });
      return;
    }

    await createGoalMethod({
      projects: selectedProjects,
      languages: selectedLanguages,
      categories: selectedCategories,
      seconds: getDurationSeconds,
      editors: selectedEditors,
      delta: targetDurationPeriod,
      target_direction: codeMore,
      type: "coding", // if this is hardcoded, what's the use?
    });
  };

  const PROJECT_OPTIONS = React.useMemo(() => {
    return projects.map((project) => {
      return { label: project.name, value: project.id || "" };
    });
  }, [projects]);

  const goalOptions: GoalOptionItem[] = [
    { label: "overall", value: "overall", id: "overall" },
    { label: "project(s)...", value: "project", id: "project" },
    { label: "language(s)...", value: "language", id: "language" },
    { label: "editor(s)...", value: "editor", id: "editor" },
    { label: "category...", value: "category", id: "category" },
  ];

  const canSpecifyDuration = React.useMemo(() => {
    const criteria = [selectedGoalOption !== ""];
    if (selectedGoalOption !== "overall") {
      criteria.push(
        [
          ...selectedProjects,
          ...selectedLanguages,
          ...selectedEditors,
          ...selectedCategories,
        ].length > 0
      );
    }
    return criteria.every((criterion) => criterion);
  }, [
    selectedGoalOption,
    selectedProjects,
    selectedLanguages,
    selectedEditors,
    selectedCategories,
  ]);

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(open) =>
        dispatch({ type: GoalActionType.SET_DIALOG_OPEN, payload: open })
      }
    >
      <DialogTrigger asChild>
        <Button
          onClick={() =>
            dispatch({ type: GoalActionType.SET_DIALOG_OPEN, payload: true })
          }
          variant="outline"
          className="bg-black text-white"
        >
          <LucidePlus className="size-4" />
          New Goal
        </Button>
      </DialogTrigger>
      <DialogContent className={cn(styles.root, "sm:max-w-[625px]")}>
        <DialogHeader className="mb-4">
          <DialogTitle className={(styles.title, styles.label)}>
            Set a Goal
          </DialogTitle>
          <DialogDescription className="sr-only">
            Create a new goal here
          </DialogDescription>
        </DialogHeader>
        {/* <hr /> */}
        <div>
          <div>
            <div className={styles.label}>I want to do</div>
          </div>
          <div className="flex gap-2">
            <div>
              <SimpleSelect
                key={selectedGoalOption}
                options={[
                  { label: "More", id: "more", value: "more" },
                  { label: "Less", id: "less", value: "less" },
                ]}
                onChange={(val) => {
                  dispatch({
                    type: GoalActionType.SET_CODE_MORE,
                    payload: val,
                  });
                }}
              />
            </div>
          </div>
        </div>
        {codeMore && (
          <div>
            <div>
              <div className={styles.label}>Work in ...</div>
            </div>
            <div className="flex gap-2">
              <div>
                {/* <LucideArrowRight className="size-4" color="#db2777" /> */}
                <SimpleSelect
                  key={selectedGoalOption}
                  options={goalOptions}
                  onChange={(val) => {
                    dispatch({
                      type: GoalActionType.SET_SELECTED_GOAL_OPTION,
                      payload: val,
                    });
                  }}
                />
              </div>
            </div>
          </div>
        )}
        {selectedGoalOption === "language" && (
          <div>
            <div>
              <div className={styles.label}>Languages...</div>
            </div>
            <div className="flex gap-2">
              <WMultiSelect
                title=""
                options={LANGUAGE_OPTIONS}
                onSelectedOptionsChanged={(options: string[]) =>
                  dispatch({
                    type: GoalActionType.SET_SELECTED_LANGUAGES,
                    payload: options,
                  })
                }
                placeholder="Select languages"
              />
            </div>
          </div>
        )}
        {selectedGoalOption === "editor" && (
          <div>
            <div>
              <div className={styles.label}>Editors...</div>
            </div>
            <div className="flex gap-2">
              <WMultiSelect
                title=""
                options={EDITOR_OPTIONS}
                onSelectedOptionsChanged={(options: string[]) =>
                  dispatch({
                    type: GoalActionType.SET_SELECTED_EDITORS,
                    payload: options,
                  })
                }
                placeholder="Select editors"
              />
            </div>
          </div>
        )}

        {selectedGoalOption === "category" && (
          <div>
            <div>
              <div className={styles.label}>Categories...</div>
            </div>
            <div className="flex gap-2">
              <WMultiSelect
                title=""
                options={CATEGORY_OPTIONS}
                onSelectedOptionsChanged={(options: string[]) =>
                  dispatch({
                    type: GoalActionType.SET_SELECTED_CATEGORIES,
                    payload: options,
                  })
                }
                placeholder="Select categories"
              />
            </div>
          </div>
        )}

        {selectedGoalOption === "project" && (
          <div>
            <div>
              <div className={styles.label}>Projects...</div>
            </div>
            <div className="flex gap-2">
              <WMultiSelect
                title=""
                options={PROJECT_OPTIONS}
                onSelectedOptionsChanged={(options: string[]) =>
                  dispatch({
                    type: GoalActionType.SET_SELECTED_PROJECTS,
                    payload: options,
                  })
                }
                placeholder="Select projects"
              />
            </div>
          </div>
        )}

        {canSpecifyDuration && (
          <div>
            <div>
              <div className={styles.label}>For...</div>
            </div>
            <div className="items-middle flex gap-2">
              <Input
                pattern="^[^eE]+$"
                onChange={(event) =>
                  dispatch({
                    type: GoalActionType.SET_TARGET_DURATION,
                    payload: event.target.value,
                  })
                }
                type="number"
                className="mr-1 inline-block w-16"
              />
              <ClickToSelect
                options={["hrs", "mins", "secs"]}
                onChange={(value: string) =>
                  dispatch({
                    type: GoalActionType.SET_TARGET_DURATION_TYPE,
                    payload: value,
                  })
                }
                value={targetDurationType}
              />
            </div>
          </div>
        )}
        <Button
          className="mt-5 block w-full"
          size={"lg"}
          style={{
            borderRadius: "6px",
            fontSize: "18px",
            padding: "10px 16px",
            lineHeight: 1.33,
            backgroundColor: "#05cfc8",
            border: "1px solid white",
          }}
          onClick={createGoal}
          disabled={
            !targetDuration || +targetDuration <= 0 || !canSpecifyDuration
          }
        >
          Set Goal
          {loading && <Icons.spinner className="mr-2 size-5 animate-spin" />}
        </Button>
        {/* BEGINNING */}
        {/* END */}
      </DialogContent>
    </Dialog>
  );
}
