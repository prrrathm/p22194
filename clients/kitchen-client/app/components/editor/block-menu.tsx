import { useEffect, useRef } from "react";
import {
  Code,
  Heading1,
  Heading2,
  Heading3,
  List,
  ListOrdered,
  Minus,
  Text,
} from "lucide-react";
import { cn } from "~/utils/cn";

export interface BlockTypeOption {
  type: string;
  label: string;
  description: string;
  icon: React.ReactNode;
}

const BLOCK_OPTIONS: BlockTypeOption[] = [
  {
    type: "rich_text",
    label: "Text",
    description: "Plain paragraph",
    icon: <Text className="h-4 w-4" />,
  },
  {
    type: "heading_1",
    label: "Heading 1",
    description: "Large section header",
    icon: <Heading1 className="h-4 w-4" />,
  },
  {
    type: "heading_2",
    label: "Heading 2",
    description: "Medium section header",
    icon: <Heading2 className="h-4 w-4" />,
  },
  {
    type: "heading_3",
    label: "Heading 3",
    description: "Small section header",
    icon: <Heading3 className="h-4 w-4" />,
  },
  {
    type: "bulleted_list",
    label: "Bulleted list",
    description: "Simple bullet list",
    icon: <List className="h-4 w-4" />,
  },
  {
    type: "numbered_list",
    label: "Numbered list",
    description: "Ordered list",
    icon: <ListOrdered className="h-4 w-4" />,
  },
  {
    type: "code",
    label: "Code",
    description: "Code snippet",
    icon: <Code className="h-4 w-4" />,
  },
  {
    type: "divider",
    label: "Divider",
    description: "Visual separator",
    icon: <Minus className="h-4 w-4" />,
  },
];

interface BlockMenuProps {
  onSelect: (type: string) => void;
  onClose: () => void;
}

export function BlockMenu({ onSelect, onClose }: BlockMenuProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    }
    document.addEventListener("keydown", handleKey);
    document.addEventListener("mousedown", handleClick);
    return () => {
      document.removeEventListener("keydown", handleKey);
      document.removeEventListener("mousedown", handleClick);
    };
  }, [onClose]);

  return (
    <div
      ref={ref}
      className="z-30 w-64 rounded-xl border border-neutral-200 bg-white shadow-xl dark:border-neutral-700 dark:bg-neutral-900 overflow-hidden"
    >
      <p className="px-3 pt-3 pb-1 text-xs font-medium text-neutral-400 dark:text-neutral-500 uppercase tracking-wide">
        Block type
      </p>
      <div className="py-1 max-h-72 overflow-y-auto">
        {BLOCK_OPTIONS.map((opt) => (
          <button
            key={opt.type}
            onClick={() => onSelect(opt.type)}
            className={cn(
              "flex w-full items-center gap-3 px-3 py-2 text-sm",
              "text-neutral-700 dark:text-neutral-300 hover:bg-neutral-50 dark:hover:bg-neutral-800",
              "transition-colors"
            )}
          >
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-neutral-200 dark:border-neutral-700 text-neutral-500 dark:text-neutral-400">
              {opt.icon}
            </span>
            <span className="flex flex-col items-start">
              <span className="font-medium text-neutral-900 dark:text-neutral-100">
                {opt.label}
              </span>
              <span className="text-xs text-neutral-400 dark:text-neutral-500">
                {opt.description}
              </span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
