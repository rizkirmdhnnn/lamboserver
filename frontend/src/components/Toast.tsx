import { useState } from "react";
import { X, ChevronDown } from "lucide-react";

export interface ToastError {
  summary: string;
  detail: string;
}

interface ToastProps {
  error: ToastError;
  onDismiss: () => void;
}

function Toast({ error, onDismiss }: ToastProps) {
  const [expanded, setExpanded] = useState(false);
  return (
    <div style={{
      position: "fixed", bottom: 24, right: 24, zIndex: 1000,
      background: "var(--bg-secondary)", border: "1px solid var(--danger)",
      borderRadius: 8, padding: "16px 20px", maxWidth: 420,
      boxShadow: "0 4px 24px rgba(0,0,0,0.5)",
    }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <span style={{ fontSize: 13, color: "var(--text-primary)", fontWeight: 500 }}>
          {error.summary}
        </span>
        <button className="btn btn-secondary btn-sm" onClick={onDismiss} style={{ flexShrink: 0 }}>
          <X size={14} />
        </button>
      </div>
      {error.detail && (
        <div style={{ marginTop: 8 }}>
          <button
            onClick={() => setExpanded(!expanded)}
            style={{ fontSize: 12, color: "var(--text-secondary)", background: "none", border: "none",
              cursor: "pointer", padding: 0, display: "flex", alignItems: "center", gap: 4 }}
          >
            <ChevronDown size={12} style={{ transform: expanded ? "rotate(180deg)" : "rotate(0deg)", transition: "transform 0.15s" }} />
            {expanded ? "Hide details" : "Show details"}
          </button>
          {expanded && (
            <pre style={{ marginTop: 6, fontSize: 11, color: "var(--text-muted)",
              background: "var(--bg-tertiary)", borderRadius: 4, padding: "8px 10px",
              overflowX: "auto", whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
              {error.detail}
            </pre>
          )}
        </div>
      )}
    </div>
  );
}

export default Toast;
