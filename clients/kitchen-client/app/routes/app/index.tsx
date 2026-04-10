import { FileText } from "lucide-react";

export default function AppIndex() {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 text-neutral-400 dark:text-neutral-600 py-24">
      <FileText className="h-10 w-10" />
      <p className="text-base font-medium">Select a page from the sidebar</p>
      <p className="text-sm">or create a new one with the + button</p>
    </div>
  );
}
