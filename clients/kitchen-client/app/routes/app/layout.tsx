import { AuthGuard } from "~/components/auth/auth-guard";
import { AppShell } from "~/components/layout/app-layout";

export default function AppLayout() {
	return (
		<AuthGuard>
			<AppShell />
		</AuthGuard>
	);
}
