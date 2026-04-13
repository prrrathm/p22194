import { useCallback, useEffect, useRef, useState } from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { getDocument, updateDocument, type Document } from "~/api/documents";
import {
	listBlocks,
	encodeContent,
	createBlock,
	decodeContent,
} from "~/api/blocks";
import { Spinner } from "~/components/ui/spinner";
import { BlockList } from "./block-list";
import type { DocumentFormValues } from "./types";

interface DocumentEditorProps {
	id: string;
	onTitleChange?: (title: string, icon?: string) => void;
}

export function DocumentEditor({ id, onTitleChange }: DocumentEditorProps) {
	const [doc, setDoc] = useState<Document | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState("");
	const titleRef = useRef<HTMLDivElement>(null);
	const titleSaveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
	const onTitleChangeRef = useRef(onTitleChange);
	onTitleChangeRef.current = onTitleChange;

	const { control, reset, setValue } = useForm<DocumentFormValues>({
		defaultValues: { title: "", blocks: [] },
	});

	const { fields, append, remove, move, insert } = useFieldArray({
		control,
		name: "blocks",
		keyName: "rhfId",
	});

	const load = useCallback(async () => {
		setLoading(true);
		setError("");
		try {
			const [fetchedDoc, fetchedBlocks] = await Promise.all([
				getDocument(id),
				listBlocks(id),
			]);
			setDoc(fetchedDoc);

			// Sort blocks by snapshot order
			const map = new Map(fetchedBlocks.map((b) => [b.id, b]));
			const ordered = fetchedDoc.snapshot
				.map((sid) => map.get(sid))
				.filter(Boolean) as typeof fetchedBlocks;
			if (ordered.length === 0) ordered.push(...fetchedBlocks);

			reset({
				title: fetchedDoc.title,
				blocks: ordered.map((b) => ({
					id: b.id,
					block_type: b.block_type,
					content: decodeContent(b.content_state ?? ""),
					content_meta: b.content_meta,
				})),
			});

			onTitleChangeRef.current?.(fetchedDoc.title, fetchedDoc.title_icon);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Failed to load document");
		} finally {
			setLoading(false);
		}
	}, [id, reset]);

	useEffect(() => {
		load();
	}, [load]);

	// Set title contentEditable content once when doc loads
	useEffect(() => {
		if (titleRef.current && doc) {
			titleRef.current.textContent = doc.title || "";
		}
	}, [doc?.id]);

	async function ensureFirstBlock() {
		if (fields.length === 0 && doc) {
			const b = await createBlock(doc.id, {
				block_type: "rich_text",
				content_state: encodeContent(""),
			});
			append({ id: b.id, block_type: "rich_text", content: "" });
		}
	}

	function handleTitleKeyDown(e: React.KeyboardEvent) {
		if (e.key === "Enter") {
			e.preventDefault();
			ensureFirstBlock();
		}
	}

	function handleTitleBlur() {
		const text = titleRef.current?.textContent ?? "";
		if (!doc) return;
		setValue("title", text);
		if (titleSaveTimer.current) clearTimeout(titleSaveTimer.current);
		titleSaveTimer.current = setTimeout(async () => {
			try {
				await updateDocument(doc.id, { title: text });
				onTitleChangeRef.current?.(text, doc.title_icon);
			} catch (err) {
				console.error("Failed to save title:", err);
			}
		}, 600);
	}

	if (loading) {
		return (
			<div className="flex h-full items-center justify-center">
				<Spinner />
			</div>
		);
	}

	if (error) {
		return (
			<div className="flex h-full items-center justify-center">
				<p className="text-sm text-red-500 dark:text-red-400">{error}</p>
			</div>
		);
	}

	return (
		<div className="mx-auto max-w-3xl w-full px-16 py-12">
			{/* Document icon */}
			{doc?.title_icon && <div className="text-5xl mb-4">{doc.title_icon}</div>}

			{/* Title */}
			<div
				ref={titleRef}
				contentEditable
				suppressContentEditableWarning
				onKeyDown={handleTitleKeyDown}
				onBlur={handleTitleBlur}
				data-placeholder="Untitled"
				className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 outline-none mb-8 leading-tight empty:before:content-[attr(data-placeholder)] empty:before:text-neutral-300 dark:empty:before:text-neutral-700"
			/>

			{/* Blocks */}
			{doc && (
				<BlockList
					documentId={doc.id}
					fields={fields}
					remove={remove}
					move={move}
					insert={insert}
				/>
			)}
		</div>
	);
}
