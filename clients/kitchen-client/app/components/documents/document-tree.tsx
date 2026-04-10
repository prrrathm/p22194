import { useEffect, useState, useCallback } from "react";
import { listDocuments, type Document } from "~/api/documents";
import { DocumentItem } from "./document-item";
import { Spinner } from "~/components/ui/spinner";

export function DocumentTree() {
  const [docs, setDocs] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);

  const load = useCallback(() => {
    setLoading(true);
    listDocuments({ status: "active" })
      .then(setDocs)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  function handleRemove(id: string) {
    setDocs((prev) => prev.filter((d) => d.id !== id));
  }

  if (loading) {
    return (
      <div className="flex justify-center py-6">
        <Spinner className="h-4 w-4" />
      </div>
    );
  }

  if (docs.length === 0) {
    return (
      <p className="px-4 py-2 text-xs text-neutral-400 dark:text-neutral-600">
        No pages yet
      </p>
    );
  }

  return (
    <div className="flex flex-col py-1">
      {docs.map((doc) => (
        <DocumentItem key={doc.id} doc={doc} onRemove={handleRemove} />
      ))}
    </div>
  );
}
