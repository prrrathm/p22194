import { type ButtonHTMLAttributes } from "react";
import { cn } from "~/utils/cn";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "ghost" | "danger";
  size?: "sm" | "md";
}

export function Button({
  variant = "primary",
  size = "md",
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      className={cn(
        "inline-flex items-center justify-center font-medium rounded-md transition-colors",
        "disabled:opacity-50 disabled:pointer-events-none",
        variant === "primary" &&
          "bg-neutral-900 text-white hover:bg-neutral-700 dark:bg-neutral-100 dark:text-neutral-900 dark:hover:bg-white",
        variant === "ghost" &&
          "text-neutral-600 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-neutral-800",
        variant === "danger" &&
          "text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/40",
        size === "sm" && "px-3 py-1.5 text-sm gap-1.5",
        size === "md" && "px-4 py-2 text-sm gap-2",
        className
      )}
    >
      {children}
    </button>
  );
}
