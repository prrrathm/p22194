import { useState } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "~/contexts/auth-context";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";

export function SignUpForm() {
	const { signUp } = useAuth();
	const navigate = useNavigate();
	const [email, setEmail] = useState("");
	const [username, setUsername] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError("");
		setLoading(true);
		try {
			await signUp(email, username, password);
			navigate("/app");
		} catch (err) {
			setError(err instanceof Error ? err.message : "Registration failed");
		} finally {
			setLoading(false);
		}
	}

	return (
		<form onSubmit={handleSubmit} className="flex flex-col gap-4">
			<Input
				id="email"
				label="Email"
				type="email"
				placeholder="you@example.com"
				value={email}
				onChange={(e) => setEmail(e.target.value)}
				required
				autoComplete="email"
			/>
			<Input
				id="username"
				label="Username"
				type="text"
				placeholder="yourname"
				value={username}
				onChange={(e) => setUsername(e.target.value)}
				required
				autoComplete="username"
			/>
			<Input
				id="password"
				label="Password"
				type="password"
				placeholder="••••••••"
				value={password}
				onChange={(e) => setPassword(e.target.value)}
				required
				autoComplete="new-password"
			/>
			{error && (
				<p className="text-sm text-red-500 dark:text-red-400">{error}</p>
			)}
			<Button type="submit" disabled={loading}>
				{loading ? "Creating account…" : "Create account"}
			</Button>
		</form>
	);
}
