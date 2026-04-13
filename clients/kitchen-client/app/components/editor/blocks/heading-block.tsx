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

interface HeadingBlockProps {
	content: string;
	level: 1 | 2 | 3;
	focused?: boolean;
	onSave: (text: string) => void;
	onBlurSave: (text: string) => void;
	onEnter: () => void;
	onBackspaceEmpty: () => void;
	onArrowUp?: () => void;
	onArrowDown?: (text: string) => void;
}

export function HeadingBlock({
	content,
	level,
	focused,
	onSave,
	onBlurSave,
	onEnter,
	onBackspaceEmpty,
	onArrowUp,
	onArrowDown,
}: HeadingBlockProps) {
	const ref = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (ref.current) {
			ref.current.textContent = content;
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

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
		} else if (
			e.key === "ArrowUp" &&
			onArrowUp &&
			ref.current &&
			isAtFirstLine(ref.current)
		) {
			e.preventDefault();
			onArrowUp();
		} else if (
			e.key === "ArrowDown" &&
			onArrowDown &&
			ref.current &&
			isAtLastLine(ref.current)
		) {
			e.preventDefault();
			onArrowDown(text);
		}
	}

	function handleBlur() {
		onBlurSave(ref.current?.textContent ?? "");
	}

	return (
		<div
			ref={ref}
			contentEditable
			suppressContentEditableWarning
			onKeyDown={handleKeyDown}
			onBlur={handleBlur}
			data-placeholder={`Heading ${level}`}
			className={cn(
				"w-full outline-none font-bold leading-tight",
				"text-neutral-900 dark:text-neutral-100",
				"empty:before:content-[attr(data-placeholder)] empty:before:text-neutral-300 dark:empty:before:text-neutral-600 empty:before:font-normal",
				level === 1 && "text-3xl mt-2",
				level === 2 && "text-xl mt-1",
				level === 3 && "text-lg mt-1",
			)}
		/>
	);
}
