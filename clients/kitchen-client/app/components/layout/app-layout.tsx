import { Outlet } from "react-router";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

interface AppLayoutProps {
  title?: string;
  icon?: string;
}

export function AppShell({ title, icon }: AppLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-white dark:bg-neutral-950">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden min-w-0">
        <Topbar title={title} icon={icon} />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
