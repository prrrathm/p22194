import { apiRequest } from "./client";

export interface Block {
	id: string;
	document_id: string;
	block_type: string;
	content_state: string; // base64-encoded bytes from the server
	content_meta?: Record<string, unknown>;
	insert_after_block_id?: string;
	deleted_at?: string;
}

export interface CreateBlockInput {
	block_type: string;
	content_state?: string; // base64
	content_meta?: Record<string, unknown>;
	insert_after_block_id?: string;
}

export interface UpdateBlockInput {
	content_state: string; // base64
	content_meta?: Record<string, unknown>;
}

/** Encode a UTF-8 string as base64 for the content_state field. */
export function encodeContent(text: string): string {
	return btoa(unescape(encodeURIComponent(text)));
}

/** Decode a base64 content_state field back to a UTF-8 string. */
export function decodeContent(base64: string): string {
	if (!base64) return "";
	try {
		return decodeURIComponent(escape(atob(base64)));
	} catch {
		return "";
	}
}

export function listBlocks(documentId: string): Promise<Block[]> {
	return apiRequest<Block[]>(`/api/v1/documents/${documentId}/blocks`);
}

export function createBlock(
	documentId: string,
	input: CreateBlockInput,
): Promise<Block> {
	return apiRequest<Block>(`/api/v1/documents/${documentId}/blocks`, {
		method: "POST",
		body: JSON.stringify(input),
	});
}

export function updateBlock(
	blockId: string,
	input: UpdateBlockInput,
): Promise<Block> {
	return apiRequest<Block>(`/api/v1/blocks/${blockId}`, {
		method: "PATCH",
		body: JSON.stringify(input),
	});
}

export function reorderBlock(
	blockId: string,
	insertAfterBlockId: string | null,
): Promise<Block> {
	return apiRequest<Block>(`/api/v1/blocks/${blockId}/reorder`, {
		method: "PATCH",
		body: JSON.stringify({ insert_after_block_id: insertAfterBlockId }),
	});
}

export function deleteBlock(blockId: string): Promise<void> {
	return apiRequest<void>(`/api/v1/blocks/${blockId}`, { method: "DELETE" });
}
