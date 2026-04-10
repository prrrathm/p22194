import { useEffect, useRef } from "react";
import { cn } from "~/utils/cn";

function isAtFirstLine(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0) return true;
  const range = sel.getRangeAt(0).cloneRange();
  range.collapse(true);
  const caret = range.getBoundingClientRect();
  if (caret.height === 0) return true;
  return caret.top < el.getBoundingClientRect().top + caret.height;
}

function isAtLastLine(el: HTMLElement): boolean {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0) return true;
  const range = sel.getRangeAt(0).cloneRange();
  range.collapse(false);
  const caret = range.getBoundingClientRect();
  if (caret.height === 0) return true;
  return caret.bottom > el.getBoundingClientRect().bottom - caret.height;
}

interface TextBlockProps {
  content: string;
  blockType: "rich_text" | "bulleted_list" | "numbered_list";
  index?: number;
  focused?: boolean;
  onSave: (text: string) => void;
  onBlurSave: (text: string) => void;
  onEnter: () => void;
  onBackspaceEmpty: () => void;
  onSlashEmpty: () => void;
  onArrowUp?: () => void;
  onArrowDown?: (text: string) => void;
}

export function TextBlock({
  content,
  blockType,
  index = 0,
  focused,
  onSave,
  onBlurSave,
  onEnter,
  onBackspaceEmpty,
  onSlashEmpty,
  onArrowUp,
  onArrowDown,
}: TextBlockProps) {
  const ref = useRef<HTMLDivElement>(null);

  // Set initial content once on mount
  useEffect(() => {
    if (ref.current) {
      ref.current.textContent = content;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Move focus when requested
  useEffect(() => {
    if (focused && ref.current) {
      ref.current.focus();
      const range = document.createRange();
      range.selectNodeContents(ref.current);
      range.collapse(false);
      window.getSelection()?.removeAllRanges();
      window.getSelection()?.addRange(range);
    }
  }, [focused]);

  function handleKeyDown(e: React.KeyboardEvent<HTMLDivElement>) {
    const text = ref.current?.textContent ?? "";
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSave(text);
      onEnter();
    } else if (e.key === "Backspace" && text === "") {
      e.preventDefault();
      onBackspaceEmpty();
    } else if (e.key === "/" && text === "") {
      onSlashEmpty();
    } else if (e.key === "ArrowUp" && onArrowUp && ref.current && isAtFirstLine(ref.current)) {
      e.preventDefault();
      onArrowUp();
    } else if (e.key === "ArrowDown" && onArrowDown && ref.current && isAtLastLine(ref.current)) {
      e.preventDefault();
      onArrowDown(text);
    }
  }

  function handleBlur() {
    onBlurSave(ref.current?.textContent ?? "");
  }

  const prefixNode =
    blockType === "bulleted_list" ? (
      <span className="shrink-0 w-4 text-center text-neutral-400 dark:text-neutral-500 select-none">
        •
      </span>
    ) : blockType === "numbered_list" ? (
      <span className="shrink-0 w-4 text-right text-neutral-400 dark:text-neutral-500 select-none text-sm">
        {index + 1}.
      </span>
    ) : null;

  return (
    <div className="flex gap-2 items-start">
      {prefixNode}
      <div
        ref={ref}
        contentEditable
        suppressContentEditableWarning
        onKeyDown={handleKeyDown}
        onBlur={handleBlur}
        data-placeholder={
          blockType === "rich_text"
            ? "Type something, or press '/' for commands…"
            : undefined
        }
        className={cn(
          "flex-1 min-h-[1.6em] outline-none leading-relaxed",
          "text-neutral-900 dark:text-neutral-100",
          "empty:before:content-[attr(data-placeholder)] empty:before:text-neutral-300 dark:empty:before:text-neutral-600",
          blockType === "rich_text" && "text-base"
        )}
      />
    </div>
  );
}
