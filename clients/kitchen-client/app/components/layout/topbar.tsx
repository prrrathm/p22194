import { Moon, Sun } from "lucide-react";
import { useTheme } from "~/contexts/theme-context";
import { Button } from "~/components/ui/button";

interface TopbarProps {
	title?: string;
	icon?: string;
}

export function Topbar({ title, icon }: TopbarProps) {
	const { theme, toggleTheme } = useTheme();

	return (
		<header className="flex h-11 shrink-0 items-center border-b border-neutral-200 dark:border-neutral-800 px-4 gap-2">
			<div className="flex-1 flex items-center gap-1.5 min-w-0">
				{icon && <span className="text-base">{icon}</span>}
				<span className="text-sm font-medium text-neutral-600 dark:text-neutral-400 truncate">
					{title ?? ""}
				</span>
			</div>
			<Button
				variant="ghost"
				size="sm"
				onClick={toggleTheme}
				className="px-2 shrink-0"
				aria-label="Toggle theme"
			>
				{theme === "dark" ? (
					<Sun className="h-4 w-4" />
				) : (
					<Moon className="h-4 w-4" />
				)}
			</Button>
		</header>
	);
}
