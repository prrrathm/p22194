import { useEffect, useRef } from "react";

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

interface CodeBlockProps {
  content: string;
  language?: string;
  focused?: boolean;
  onSave: (text: string) => void;
  onBlurSave: (text: string) => void;
  onEnter: () => void;
  onBackspaceEmpty: () => void;
  onArrowUp?: () => void;
  onArrowDown?: (text: string) => void;
}

export function CodeBlock({
  content,
  language,
  focused,
  onSave,
  onBlurSave,
  onEnter,
  onBackspaceEmpty,
  onArrowUp,
  onArrowDown,
}: CodeBlockProps) {
  const ref = useRef<HTMLElement>(null);

  useEffect(() => {
    if (ref.current) {
      ref.current.textContent = content;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (focused && ref.current) {
      ref.current.focus();
    }
  }, [focused]);

  function handleKeyDown(e: React.KeyboardEvent<HTMLElement>) {
    const text = ref.current?.textContent ?? "";
    if (e.key === "Enter" && e.shiftKey) {
      e.preventDefault();
      onSave(text);
      onEnter();
    } else if (e.key === "Backspace" && text === "") {
      e.preventDefault();
      onBackspaceEmpty();
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

  return (
    <div className="rounded-lg border border-neutral-200 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 overflow-hidden">
      {language && (
        <div className="px-4 py-1.5 border-b border-neutral-200 dark:border-neutral-700 text-xs text-neutral-400 dark:text-neutral-500 font-mono">
          {language}
        </div>
      )}
      <pre className="overflow-x-auto">
        <code
          ref={ref}
          contentEditable
          suppressContentEditableWarning
          onKeyDown={handleKeyDown}
          onBlur={handleBlur}
          className="block px-4 py-3 text-sm font-mono text-neutral-800 dark:text-neutral-200 outline-none whitespace-pre leading-relaxed"
          data-placeholder="// code here…"
        />
      </pre>
      <p className="px-4 pb-1.5 text-xs text-neutral-400 dark:text-neutral-600">
        Shift+Enter to exit block
      </p>
    </div>
  );
}
