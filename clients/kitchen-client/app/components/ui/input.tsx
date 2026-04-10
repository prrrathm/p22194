import { type InputHTMLAttributes } from "react";
import { cn } from "~/utils/cn";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export function Input({ label, error, className, id, ...props }: InputProps) {
  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label
          htmlFor={id}
          className="text-sm font-medium text-neutral-700 dark:text-neutral-300"
        >
          {label}
        </label>
      )}
      <input
        id={id}
        {...props}
        className={cn(
          "w-full rounded-md border bg-white px-3 py-2 text-sm text-neutral-900 outline-none",
          "placeholder:text-neutral-400 transition-colors",
          "border-neutral-200 focus:border-neutral-400 focus:ring-1 focus:ring-neutral-400",
          "dark:bg-neutral-900 dark:border-neutral-700 dark:text-neutral-100",
          "dark:placeholder:text-neutral-500 dark:focus:border-neutral-500 dark:focus:ring-neutral-500",
          error && "border-red-400 focus:border-red-400 focus:ring-red-400",
          className
        )}
      />
      {error && (
        <p className="text-xs text-red-500 dark:text-red-400">{error}</p>
      )}
    </div>
  );
}
