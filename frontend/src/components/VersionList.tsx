import { Check, Trash2 } from "lucide-react";

interface InstalledVersion {
  version: string;
  path: string;
  active: boolean;
  source?: string;
}

interface VersionListProps {
  versions: InstalledVersion[];
  label: string;
  showSource?: boolean;
  canUninstall?: (v: InstalledVersion) => boolean;
  onSetActive: (version: string) => void;
  onUninstall?: (version: string) => void;
}

function VersionList({
  versions,
  label,
  showSource,
  canUninstall,
  onSetActive,
  onUninstall,
}: VersionListProps) {
  return (
    <>
      {versions.map((v) => (
        <div
          key={v.version}
          className={`version-item ${v.active ? "active" : ""}`}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <div>
              <div style={{ fontWeight: 600, fontSize: 14 }}>
                {label} {v.version}
              </div>
              <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
                {v.path}
              </div>
            </div>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            {showSource && v.source && (
              <span className="version-source">{v.source}</span>
            )}
            {v.active ? (
              <span className="version-badge">
                <Check size={12} style={{ marginRight: 4 }} />
                Active
              </span>
            ) : (
              <>
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => onSetActive(v.version)}
                >
                  Use
                </button>
                {onUninstall &&
                  canUninstall &&
                  canUninstall(v) && (
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => onUninstall(v.version)}
                    >
                      <Trash2 size={14} />
                    </button>
                  )}
              </>
            )}
          </div>
        </div>
      ))}
    </>
  );
}

export default VersionList;
