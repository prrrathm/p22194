import { useState } from "react";
import { useParams } from "react-router";
import { DocumentEditor } from "~/components/editor/document-editor";

export default function DocumentPage() {
  const { id } = useParams<{ id: string }>();
  // Title/icon lifted up so the topbar in AppShell can receive them
  // For now the topbar is static; title updates are visible in the editor itself
  const [, setMeta] = useState({ title: "", icon: "" });

  if (!id) return null;

  return (
    <DocumentEditor
      id={id}
      onTitleChange={(title, icon) => setMeta({ title, icon: icon ?? "" })}
    />
  );
}
