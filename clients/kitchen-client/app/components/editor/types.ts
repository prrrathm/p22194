export interface BlockField {
	id: string;
	block_type: string;
	content: string;
	content_meta?: Record<string, unknown>;
}

export interface DocumentFormValues {
	title: string;
	blocks: BlockField[];
}
