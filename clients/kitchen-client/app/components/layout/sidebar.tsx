import { LogOut, Plus } from "lucide-react";
import { useNavigate } from "react-router";
import { useAuth } from "~/contexts/auth-context";
import { createDocument } from "~/api/documents";
import { DocumentTree } from "~/components/documents/document-tree";
import { Button } from "~/components/ui/button";

export function Sidebar() {
  const { user, signOut } = useAuth();
  const navigate = useNavigate();

  async function handleNewPage() {
    try {
      const doc = await createDocument({ title: "Untitled" });
      navigate(`/app/doc/${doc.id}`);
    } catch (err) {
      console.error("Failed to create document:", err);
    }
  }

  return (
    <aside className="flex flex-col w-60 shrink-0 border-r border-neutral-200 dark:border-neutral-800 bg-neutral-50 dark:bg-neutral-950 h-screen overflow-hidden">
      {/* Workspace header */}
      <div className="flex items-center gap-2 px-3 h-11 shrink-0 border-b border-neutral-200 dark:border-neutral-800">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-neutral-900 dark:text-neutral-100 truncate">
            {user?.email ?? ""}
          </p>
        </div>
      </div>

      {/* New page */}
      <div className="px-2 pt-2 shrink-0">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleNewPage}
          className="w-full justify-start text-neutral-500 dark:text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300"
        >
          <Plus className="h-4 w-4" />
          New page
        </Button>
      </div>

      {/* Document list */}
      <div className="flex-1 overflow-y-auto py-1 min-h-0">
        <DocumentTree />
      </div>

      {/* Sign out */}
      <div className="px-2 py-2 shrink-0 border-t border-neutral-200 dark:border-neutral-800">
        <Button
          variant="ghost"
          size="sm"
          onClick={signOut}
          className="w-full justify-start text-neutral-500 dark:text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </Button>
      </div>
    </aside>
  );
}
