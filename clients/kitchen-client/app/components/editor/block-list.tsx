import { useCallback, useRef, useState } from "react";
import { GripVertical } from "lucide-react";
import type {
	UseFieldArrayInsert,
	UseFieldArrayMove,
	UseFieldArrayRemove,
	FieldArrayWithId,
} from "react-hook-form";
import {
	createBlock,
	deleteBlock,
	encodeContent,
	reorderBlock,
	updateBlock,
} from "~/api/blocks";
import { cn } from "~/utils/cn";
import { BlockRenderer } from "./block";
import { BlockMenu } from "./block-menu";
import type { BlockField, DocumentFormValues } from "./types";

interface BlockListProps {
	documentId: string;
	fields: FieldArrayWithId<DocumentFormValues, "blocks", "rhfId">[];
	remove: UseFieldArrayRemove;
	move: UseFieldArrayMove;
	insert: UseFieldArrayInsert<DocumentFormValues, "blocks">;
}

export function BlockList({
	documentId,
	fields,
	remove,
	move,
	insert,
}: BlockListProps) {
	const [focusedId, setFocusedId] = useState<string | null>(null);
	const [menuFor, setMenuFor] = useState<string | null>(null);
	const [handleMenuFor, setHandleMenuFor] = useState<string | null>(null);

	// Drag-and-drop state
	const [draggedId, setDraggedId] = useState<string | null>(null);
	const [dragOverId, setDragOverId] = useState<string | null>(null);
	const [dragPosition, setDragPosition] = useState<"above" | "below">("below");

	const saveTimers = useRef<Map<string, ReturnType<typeof setTimeout>>>(
		new Map(),
	);

	// Debounced save — avoids hammering the API on every keystroke
	const scheduleSave = useCallback((blockId: string, text: string) => {
		const existing = saveTimers.current.get(blockId);
		if (existing) clearTimeout(existing);
		const timer = setTimeout(async () => {
			try {
				await updateBlock(blockId, { content_state: encodeContent(text) });
			} catch (err) {
				console.error("Failed to save block:", err);
			}
			saveTimers.current.delete(blockId);
		}, 100);
		saveTimers.current.set(blockId, timer);
	}, []);

	// Immediate save — cancels any pending debounce and persists right away.
	const flushSave = useCallback((blockId: string, text: string) => {
		const existing = saveTimers.current.get(blockId);
		if (existing) clearTimeout(existing);
		saveTimers.current.delete(blockId);
		updateBlock(blockId, { content_state: encodeContent(text) }).catch(
			(err) => {
				console.error("Failed to save block:", err);
			},
		);
	}, []);

	async function handleEnter(afterBlockId: string) {
		try {
			const newBlock = await createBlock(documentId, {
				block_type: "rich_text",
				content_state: encodeContent(""),
				insert_after_block_id: afterBlockId,
			});
			const idx = fields.findIndex((f) => f.id === afterBlockId);
			insert(idx + 1, {
				id: newBlock.id,
				block_type: "rich_text",
				content: "",
			});
			setFocusedId(newBlock.id);
		} catch (err) {
			console.error("Failed to create block:", err);
		}
	}

	async function handleDelete(blockId: string) {
		if (fields.length <= 1) return;
		try {
			await deleteBlock(blockId);
			const idx = fields.findIndex((f) => f.id === blockId);
			remove(idx);
			const focusTarget = fields[Math.max(0, idx - 1)];
			if (focusTarget) setFocusedId(focusTarget.id);
		} catch (err) {
			console.error("Failed to delete block:", err);
		}
	}

	// ── Arrow key navigation ──────────────────────────────────────────────────

	function handleArrowUp(blockId: string) {
		const idx = fields.findIndex((f) => f.id === blockId);
		if (idx > 0) setFocusedId(fields[idx - 1].id);
	}

	async function handleArrowDown(blockId: string, currentText: string) {
		const idx = fields.findIndex((f) => f.id === blockId);
		if (idx < fields.length - 1) {
			setFocusedId(fields[idx + 1].id);
		} else if (currentText.trim() !== "") {
			await handleEnter(blockId);
		}
	}

	// ── Slash-menu block type change ──────────────────────────────────────────

	async function handleBlockTypeSelect(blockId: string, type: string) {
		setMenuFor(null);
		setHandleMenuFor(null);
		try {
			const newBlock = await createBlock(documentId, {
				block_type: type,
				content_state: encodeContent(""),
				insert_after_block_id: blockId,
			});
			const idx = fields.findIndex((f) => f.id === blockId);
			await deleteBlock(blockId);
			remove(idx);
			insert(idx, {
				id: newBlock.id,
				block_type: type,
				content: "",
				content_meta: newBlock.content_meta,
			});
			setFocusedId(newBlock.id);
		} catch (err) {
			console.error("Failed to change block type:", err);
		}
	}

	// ── Drag-and-drop ─────────────────────────────────────────────────────────

	function handleDragStart(e: React.DragEvent, blockId: string) {
		setDraggedId(blockId);
		e.dataTransfer.effectAllowed = "move";
	}

	function handleDragOver(e: React.DragEvent, blockId: string) {
		e.preventDefault();
		e.dataTransfer.dropEffect = "move";
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		setDragOverId(blockId);
		setDragPosition(e.clientY < rect.top + rect.height / 2 ? "above" : "below");
	}

	function handleDragLeave(e: React.DragEvent, blockId: string) {
		const related = e.relatedTarget as Node | null;
		if (related && (e.currentTarget as HTMLElement).contains(related)) return;
		if (dragOverId === blockId) setDragOverId(null);
	}

	async function handleDrop(e: React.DragEvent, targetBlockId: string) {
		e.preventDefault();
		if (!draggedId || draggedId === targetBlockId) {
			setDraggedId(null);
			setDragOverId(null);
			return;
		}

		const fromIdx = fields.findIndex((f) => f.id === draggedId);
		const toTargetIdx = fields.findIndex((f) => f.id === targetBlockId);

		// Compute insert position in the post-removal array
		const insertIdx = dragPosition === "above" ? toTargetIdx : toTargetIdx + 1;
		const finalIdx = insertIdx > fromIdx ? insertIdx - 1 : insertIdx;

		// insertAfterBlockId: block preceding the moved block after reorder, or null for top
		const tempFields = [...fields];
		const [moved] = tempFields.splice(fromIdx, 1);
		tempFields.splice(finalIdx, 0, moved);
		const insertAfterBlockId =
			finalIdx === 0 ? null : tempFields[finalIdx - 1].id;

		// Optimistic update
		move(fromIdx, finalIdx);
		setDraggedId(null);
		setDragOverId(null);

		try {
			await reorderBlock(draggedId, insertAfterBlockId);
		} catch (err) {
			console.error("Failed to reorder block:", err);
			// Rollback
			move(finalIdx, fromIdx);
		}
	}

	function handleDragEnd() {
		setDraggedId(null);
		setDragOverId(null);
	}

	// Build a Block-like object from a field for BlockRenderer
	function fieldToBlock(
		field: FieldArrayWithId<DocumentFormValues, "blocks", "rhfId">,
	): BlockField & { document_id: string; content_state: string } {
		return {
			...field,
			document_id: documentId,
			content_state: encodeContent(field.content),
		};
	}

	return (
		<div className="relative">
			<div className="flex flex-col gap-0.5 ml-6">
				{fields.map((field, i) => (
					<div
						key={field.rhfId}
						className={cn(
							"relative group px-1 py-0.5 rounded-md",
							"hover:bg-neutral-50 dark:hover:bg-neutral-900/50",
							draggedId === field.id && "opacity-40",
						)}
						onDragOver={(e) => handleDragOver(e, field.id)}
						onDragLeave={(e) => handleDragLeave(e, field.id)}
						onDrop={(e) => handleDrop(e, field.id)}
					>
						{/* Drop indicator — above */}
						{dragOverId === field.id && dragPosition === "above" && (
							<div className="pointer-events-none absolute -top-px left-0 right-0 h-0.5 rounded bg-blue-500" />
						)}

						{/* Drag handle — visible on hover */}
						<div className="absolute -left-6 inset-y-0 flex items-center opacity-0 group-hover:opacity-100 transition-opacity">
							<button
								draggable
								onDragStart={(e) => handleDragStart(e, field.id)}
								onDragEnd={handleDragEnd}
								onClick={(e) => {
									e.stopPropagation();
									setHandleMenuFor(
										handleMenuFor === field.id ? null : field.id,
									);
								}}
								className={cn(
									"p-0.5 rounded text-neutral-300 dark:text-neutral-600",
									"hover:text-neutral-500 dark:hover:text-neutral-400",
									"hover:bg-neutral-100 dark:hover:bg-neutral-800",
									"cursor-grab active:cursor-grabbing",
								)}
								title="Drag to reorder, click for options"
							>
								<GripVertical className="h-4 w-4" />
							</button>
						</div>

						{/* Handle-click menu (block type / "Turn into") */}
						{handleMenuFor === field.id && (
							<div className="absolute left-0 top-full mt-1 z-30">
								<BlockMenu
									onSelect={(type) => handleBlockTypeSelect(field.id, type)}
									onClose={() => setHandleMenuFor(null)}
								/>
							</div>
						)}

						<BlockRenderer
							block={fieldToBlock(field)}
							index={i}
							focused={focusedId === field.id}
							onSave={scheduleSave}
							onFlushSave={flushSave}
							onEnter={handleEnter}
							onDelete={handleDelete}
							onSlashEmpty={(id) => setMenuFor(id)}
							onArrowUp={handleArrowUp}
							onArrowDown={handleArrowDown}
						/>

						{/* Slash-command menu */}
						{menuFor === field.id && (
							<div className="absolute left-0 top-full mt-1 z-30">
								<BlockMenu
									onSelect={(type) => handleBlockTypeSelect(field.id, type)}
									onClose={() => setMenuFor(null)}
								/>
							</div>
						)}

						{/* Drop indicator — below */}
						{dragOverId === field.id && dragPosition === "below" && (
							<div className="pointer-events-none absolute -bottom-px left-0 right-0 h-0.5 rounded bg-blue-500" />
						)}
					</div>
				))}
			</div>

			{/* Click below last block to add a new one */}
			{fields.length > 0 && (
				<button
					onClick={async () => {
						const last = fields[fields.length - 1];
						await handleEnter(last.id);
					}}
					className="w-full mt-1 py-4 text-left text-sm text-neutral-300 dark:text-neutral-700 hover:text-neutral-400 dark:hover:text-neutral-500 transition-colors pl-7"
				>
					Click to add a block…
				</button>
			)}
		</div>
	);
}
