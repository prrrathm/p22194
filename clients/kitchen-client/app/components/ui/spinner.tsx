import { cn } from "~/utils/cn";

export function Spinner({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "h-5 w-5 animate-spin rounded-full border-2",
        "border-neutral-200 border-t-neutral-600",
        "dark:border-neutral-700 dark:border-t-neutral-300",
        className
      )}
    />
  );
}
