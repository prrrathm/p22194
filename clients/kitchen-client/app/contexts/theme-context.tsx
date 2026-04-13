import {
	createContext,
	useContext,
	useEffect,
	useState,
	type ReactNode,
} from "react";

type Theme = "light" | "dark";

interface ThemeContextValue {
	theme: Theme;
	toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function resolveInitialTheme(): Theme {
	if (typeof window === "undefined") return "light";
	const stored = localStorage.getItem("theme");
	if (stored === "dark" || stored === "light") return stored;
	return window.matchMedia("(prefers-color-scheme: dark)").matches
		? "dark"
		: "light";
}

export function ThemeProvider({ children }: { children: ReactNode }) {
	const [theme, setTheme] = useState<Theme>("light");

	useEffect(() => {
		const initial = resolveInitialTheme();
		setTheme(initial);
		document.documentElement.classList.toggle("dark", initial === "dark");
	}, []);

	function toggleTheme() {
		setTheme((prev) => {
			const next: Theme = prev === "dark" ? "light" : "dark";
			localStorage.setItem("theme", next);
			document.documentElement.classList.toggle("dark", next === "dark");
			return next;
		});
	}

	return (
		<ThemeContext.Provider value={{ theme, toggleTheme }}>
			{children}
		</ThemeContext.Provider>
	);
}

export function useTheme(): ThemeContextValue {
	const ctx = useContext(ThemeContext);
	if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
	return ctx;
}
