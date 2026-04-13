import { type Block } from "~/api/blocks";
import { decodeContent } from "~/api/blocks";
import { TextBlock } from "./blocks/text-block";
import { HeadingBlock } from "./blocks/heading-block";
import { CodeBlock } from "./blocks/code-block";
import { DividerBlock } from "./blocks/divider-block";

interface BlockProps {
	block: Block;
	index: number;
	focused?: boolean;
	onSave: (blockId: string, text: string) => void;
	onFlushSave: (blockId: string, text: string) => void;
	onEnter: (blockId: string) => void;
	onDelete: (blockId: string) => void;
	onSlashEmpty: (blockId: string) => void;
	onArrowUp: (blockId: string) => void;
	onArrowDown: (blockId: string, text: string) => void;
}

export function BlockRenderer({
	block,
	index,
	focused,
	onSave,
	onFlushSave,
	onEnter,
	onDelete,
	onSlashEmpty,
	onArrowUp,
	onArrowDown,
}: BlockProps) {
	const content = decodeContent(block.content_state ?? "");
	const language = block.content_meta?.language as string | undefined;

	const sharedProps = {
		focused,
		onSave: (text: string) => onSave(block.id, text),
		onBlurSave: (text: string) => onFlushSave(block.id, text),
		onEnter: () => onEnter(block.id),
		onBackspaceEmpty: () => onDelete(block.id),
		onArrowUp: () => onArrowUp(block.id),
		onArrowDown: (text: string) => onArrowDown(block.id, text),
	};

	switch (block.block_type) {
		case "heading_1":
			return <HeadingBlock {...sharedProps} content={content} level={1} />;
		case "heading_2":
			return <HeadingBlock {...sharedProps} content={content} level={2} />;
		case "heading_3":
			return <HeadingBlock {...sharedProps} content={content} level={3} />;
		case "code":
			return (
				<CodeBlock
					{...sharedProps}
					content={content}
					language={language}
					onEnter={() => onEnter(block.id)}
				/>
			);
		case "divider":
			return <DividerBlock onDelete={() => onDelete(block.id)} />;
		case "bulleted_list":
			return (
				<TextBlock
					{...sharedProps}
					content={content}
					blockType="bulleted_list"
					index={index}
					onSlashEmpty={() => onSlashEmpty(block.id)}
				/>
			);
		case "numbered_list":
			return (
				<TextBlock
					{...sharedProps}
					content={content}
					blockType="numbered_list"
					index={index}
					onSlashEmpty={() => onSlashEmpty(block.id)}
				/>
			);
		default:
			return (
				<TextBlock
					{...sharedProps}
					content={content}
					blockType="rich_text"
					onSlashEmpty={() => onSlashEmpty(block.id)}
				/>
			);
	}
}
