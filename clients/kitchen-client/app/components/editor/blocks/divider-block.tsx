interface DividerBlockProps {
	onDelete: () => void;
}

export function DividerBlock({ onDelete }: DividerBlockProps) {
	return (
		<div
			className="group relative flex items-center cursor-pointer py-1"
			onClick={onDelete}
			title="Click to delete"
		>
			<hr className="w-full border-neutral-200 dark:border-neutral-700" />
			<span className="absolute right-0 opacity-0 group-hover:opacity-100 text-xs text-neutral-400 dark:text-neutral-600 bg-white dark:bg-neutral-950 px-1 transition-opacity">
				delete
			</span>
		</div>
	);
}
