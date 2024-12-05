import React from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

import { Icons } from "../icons";

interface AlertDialogDemoProps {
  children: React.ReactNode;
  description: string;
  title: string;
  onCancel?: () => void;
  onConfirm: () => void;
  open?: boolean;
  loading?: boolean;
  cancelText?: string;
  continueText?: string;
  continueClassName?: string;
  titleClassName?: string;
}

export function Confirm(props: AlertDialogDemoProps) {
  const topProps: { open?: boolean } = {};
  const {
    cancelText = "Cancel",
    continueText = "Continue",
    continueClassName = "",
    titleClassName = "",
  } = props;
  if (props.open) {
    topProps.open = props.open;
  }
  return (
    <AlertDialog {...topProps}>
      <AlertDialogTrigger asChild>{props.children}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle className={titleClassName}>
            {props.title}
          </AlertDialogTitle>
          <AlertDialogDescription>{props.description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={props.onCancel}>
            {cancelText}
          </AlertDialogCancel>
          <AlertDialogAction
            className={continueClassName}
            onClick={props.onConfirm}
          >
            {props.loading && (
              <Icons.spinner className="mr-2 size-4 animate-spin" />
            )}
            {continueText}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
