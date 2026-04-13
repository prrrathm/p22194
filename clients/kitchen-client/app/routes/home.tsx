import { useEffect } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "~/contexts/auth-context";
import { Spinner } from "~/components/ui/spinner";

export function meta() {
	return [{ title: "Kitchen" }];
}

export default function Home() {
	const { user, isLoading } = useAuth();
	const navigate = useNavigate();

	useEffect(() => {
		if (!isLoading) {
			navigate(user ? "/app" : "/auth/sign-in", { replace: true });
		}
	}, [user, isLoading, navigate]);

	return (
		<div className="flex h-screen items-center justify-center bg-white dark:bg-neutral-950">
			<Spinner />
		</div>
	);
}
