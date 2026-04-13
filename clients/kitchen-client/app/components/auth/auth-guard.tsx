import { useEffect, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "~/contexts/auth-context";
import { Spinner } from "~/components/ui/spinner";

export function AuthGuard({ children }: { children: ReactNode }) {
	const { user, isLoading } = useAuth();
	const navigate = useNavigate();

	useEffect(() => {
		if (!isLoading && !user) {
			navigate("/auth/sign-in", { replace: true });
		}
	}, [user, isLoading, navigate]);

	if (isLoading) {
		return (
			<div className="flex h-screen items-center justify-center bg-white dark:bg-neutral-950">
				<Spinner />
			</div>
		);
	}

	if (!user) return null;

	return <>{children}</>;
}
