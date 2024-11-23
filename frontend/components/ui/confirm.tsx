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
}

export function Confirm(props: AlertDialogDemoProps) {
  const topProps: { open?: boolean } = {};
  if (props.open) {
    // hack alert
    topProps.open = props.open;
  }
  return (
    <AlertDialog {...topProps}>
      <AlertDialogTrigger asChild>{props.children}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{props.title}</AlertDialogTitle>
          <AlertDialogDescription>{props.description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={props.onCancel}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={props.onConfirm}>
            {props.loading && (
              <Icons.spinner className="mr-2 size-4 animate-spin" />
            )}
            Continue
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
