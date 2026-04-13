import { apiRequest } from "./client";

export interface Document {
	id: string;
	created_by_user_id: string;
	created_at: string;
	last_updated_at: string;
	title: string;
	title_icon?: string;
	status: string;
	parent_document_id?: string;
	snapshot: string[];
}

export interface CreateDocumentInput {
	title?: string;
	title_icon?: string;
	parent_document_id?: string;
}

export interface UpdateDocumentInput {
	title?: string;
	title_icon?: string;
}

export function createDocument(
	input: CreateDocumentInput = {},
): Promise<Document> {
	return apiRequest<Document>("/api/v1/documents", {
		method: "POST",
		body: JSON.stringify(input),
	});
}

export async function listDocuments(params?: {
	status?: string;
	parent_id?: string;
}): Promise<Document[]> {
	const qs = new URLSearchParams();
	if (params?.status) qs.set("status", params.status);
	if (params?.parent_id) qs.set("parent_id", params.parent_id);
	const query = qs.toString();
	const res = await apiRequest<{ documents: Document[] }>(
		`/api/v1/documents${query ? `?${query}` : ""}`,
	);
	return res.documents ?? [];
}

export function getDocument(id: string): Promise<Document> {
	return apiRequest<Document>(`/api/v1/documents/${id}`);
}

export async function listChildren(id: string): Promise<Document[]> {
	const res = await apiRequest<{ documents: Document[] }>(
		`/api/v1/documents/${id}/children`,
	);
	return res.documents ?? [];
}

export function updateDocument(
	id: string,
	input: UpdateDocumentInput,
): Promise<Document> {
	return apiRequest<Document>(`/api/v1/documents/${id}`, {
		method: "PATCH",
		body: JSON.stringify(input),
	});
}

export function deleteDocument(id: string): Promise<void> {
	return apiRequest<void>(`/api/v1/documents/${id}`, { method: "DELETE" });
}

export function archiveDocument(id: string): Promise<void> {
	return apiRequest<void>(`/api/v1/documents/${id}/archive`, {
		method: "PATCH",
	});
}
