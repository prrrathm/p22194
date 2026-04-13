import { useRef, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { Archive, FileText, MoreHorizontal, Trash2 } from "lucide-react";
import {
	archiveDocument,
	deleteDocument,
	type Document,
} from "~/api/documents";
import { cn } from "~/utils/cn";

interface DocumentItemProps {
	doc: Document;
	onRemove: (id: string) => void;
}

export function DocumentItem({ doc, onRemove }: DocumentItemProps) {
	const navigate = useNavigate();
	const params = useParams();
	const [menuOpen, setMenuOpen] = useState(false);
	const menuRef = useRef<HTMLDivElement>(null);
	const isActive = params.id === doc.id;

	async function handleDelete(e: React.MouseEvent) {
		e.stopPropagation();
		setMenuOpen(false);
		try {
			await deleteDocument(doc.id);
			onRemove(doc.id);
			if (isActive) navigate("/app");
		} catch (err) {
			console.error(err);
		}
	}

	async function handleArchive(e: React.MouseEvent) {
		e.stopPropagation();
		setMenuOpen(false);
		try {
			await archiveDocument(doc.id);
			onRemove(doc.id);
			if (isActive) navigate("/app");
		} catch (err) {
			console.error(err);
		}
	}

	return (
		<div
			role="button"
			tabIndex={0}
			onKeyDown={(e) => e.key === "Enter" && navigate(`/app/doc/${doc.id}`)}
			onClick={() => navigate(`/app/doc/${doc.id}`)}
			className={cn(
				"group relative flex items-center gap-1.5 px-2 py-1 mx-1 rounded-md cursor-pointer select-none",
				"text-sm text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800/60",
				isActive &&
					"bg-neutral-100 dark:bg-neutral-800 text-neutral-900 dark:text-neutral-100 font-medium",
			)}
		>
			<FileText className="h-3.5 w-3.5 shrink-0 text-neutral-400 dark:text-neutral-500" />
			<span className="flex-1 truncate">{doc.title || "Untitled"}</span>

			<button
				onClick={(e) => {
					e.stopPropagation();
					setMenuOpen((v) => !v);
				}}
				className="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-opacity"
				aria-label="More options"
			>
				<MoreHorizontal className="h-3.5 w-3.5" />
			</button>

			{menuOpen && (
				<>
					<div
						className="fixed inset-0 z-10"
						onClick={(e) => {
							e.stopPropagation();
							setMenuOpen(false);
						}}
					/>
					<div
						ref={menuRef}
						className="absolute right-0 top-full z-20 mt-1 w-38 rounded-lg border border-neutral-200 bg-white shadow-lg dark:border-neutral-700 dark:bg-neutral-900"
					>
						<button
							onClick={handleArchive}
							className="flex w-full items-center gap-2 px-3 py-2 text-xs text-neutral-600 hover:bg-neutral-50 dark:text-neutral-400 dark:hover:bg-neutral-800 rounded-t-lg"
						>
							<Archive className="h-3.5 w-3.5" />
							Archive
						</button>
						<button
							onClick={handleDelete}
							className="flex w-full items-center gap-2 px-3 py-2 text-xs text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/40 rounded-b-lg"
						>
							<Trash2 className="h-3.5 w-3.5" />
							Delete
						</button>
					</div>
				</>
			)}
		</div>
	);
}
