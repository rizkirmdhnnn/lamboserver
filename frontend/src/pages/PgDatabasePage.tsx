import { useEffect, useState } from "react";
import { Database, Trash2, Plus, RefreshCw } from "lucide-react";
import Toast, { ToastError } from "../components/Toast";
import {
  StartService,
  StopService,
  GetAllStatuses,
  InstallVersion,
  InitService,
  ListServiceDatabases,
  CreateServiceDatabase,
  DropServiceDatabase,
  GetPgwebStatus,
  InstallPgweb,
  StartPgweb,
  StopPgweb,
  OpenPgweb,
} from "../../wailsjs/go/main/App";

interface PgDatabase {
  name: string;
  size: string;
}

function PgDatabasePage() {
  const [installed, setInstalled] = useState(false);
  const [running, setRunning] = useState(false);
  const [databases, setDatabases] = useState<PgDatabase[]>([]);
  const [newDbName, setNewDbName] = useState("");
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [installStep, setInstallStep] = useState<string | null>(null);
  const [toastError, setToastError] = useState<ToastError | null>(null);
  const [dropTarget, setDropTarget] = useState<string | null>(null);
  const [dropping, setDropping] = useState(false);
  const [statusLoaded, setStatusLoaded] = useState(false);
  const [pgwebInstalled, setPgwebInstalled] = useState(false);
  const [pgwebRunning, setPgwebRunning] = useState(false);
  const [pgwebLoading, setPgwebLoading] = useState(false);
  const [pgwebInstallStep, setPgwebInstallStep] = useState<string | null>(null);
  const [pgwebError, setPgwebError] = useState<string | null>(null);

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      const statuses = await GetAllStatuses();
      const pgStatus = statuses["postgresql"] || "not_installed";
      const isRunning = pgStatus === "running";
      setRunning(isRunning);
      setInstalled(pgStatus !== "not_installed");

      // Load pgweb status independently (dedicated IPC call)
      const pgwStatus = await GetPgwebStatus();
      setPgwebInstalled(pgwStatus.installed);
      setPgwebRunning(pgwStatus.running);

      setStatusLoaded(true);
      if (isRunning) {
        await loadDatabases();
      } else {
        setDatabases([]);
      }
    } catch (e) {
      console.error("Failed to load status:", e);
      setStatusLoaded(true);
    }
  };

  const loadDatabases = async () => {
    try {
      const dbs = await ListServiceDatabases("postgresql");
      setDatabases((dbs || []).map((d: any) => ({ name: d.name, size: d.size || "" })));
    } catch (e) {
      console.error("Failed to list databases:", e);
      setDatabases([]);
    }
  };

  const parseInstallError = (raw: string): ToastError => {
    const cleaned = raw.replace(/^Error invoking method "[^"]+": /, "");
    const firstLine = cleaned.split("\n")[0].slice(0, 120);
    return { summary: firstLine, detail: cleaned };
  };

  const handleInstall = async () => {
    setLoading(true);
    setToastError(null);
    try {
      setInstallStep("Downloading...");
      await InstallVersion("postgresql", "");
      setInstallStep("Initializing...");
      await InitService("postgresql");
      setInstallStep("\u2713 Installed");
      setTimeout(async () => {
        setInstallStep(null);
        setLoading(false);
        await loadStatus();
      }, 1200);
      return;
    } catch (e: any) {
      console.error(e);
      setInstallStep(null);
      setToastError(parseInstallError(String(e)));
    }
    setLoading(false);
  };

  const handleStart = async () => {
    setLoading(true);
    setError(null);
    try {
      await StartService("postgresql");
      await loadStatus();
    } catch (e: any) {
      console.error(e);
      setError(String(e));
    }
    setLoading(false);
  };

  const handleStop = async () => {
    setLoading(true);
    setError(null);
    try {
      await StopService("postgresql");
      await loadStatus();
      setDatabases([]);
    } catch (e: any) {
      console.error(e);
      setError(String(e));
    }
    setLoading(false);
  };

  const handleCreate = async () => {
    if (!newDbName.trim()) return;
    setCreating(true);
    setError(null);
    try {
      await CreateServiceDatabase("postgresql", newDbName.trim());
      setNewDbName("");
      await loadDatabases();
    } catch (e: any) {
      console.error(e);
      setError(String(e));
    }
    setCreating(false);
  };

  const handleDrop = async (name: string) => {
    setDropTarget(name);
  };

  const confirmDrop = async () => {
    if (!dropTarget) return;
    setDropping(true);
    setError(null);
    try {
      await DropServiceDatabase("postgresql", dropTarget);
      await loadDatabases();
    } catch (e: any) {
      console.error(e);
      setError(String(e));
    }
    setDropping(false);
    setDropTarget(null);
  };

  // ── pgweb handlers ────────────────────────────────────────────

  const handleInstallPgweb = async () => {
    setPgwebLoading(true);
    setPgwebError(null);
    try {
      setPgwebInstallStep("Downloading...");
      await InstallPgweb();
      setPgwebInstallStep("\u2713 Installed");
      setTimeout(async () => {
        setPgwebInstallStep(null);
        setPgwebLoading(false);
        await loadStatus();
      }, 1200);
      return;
    } catch (e: any) {
      console.error(e);
      setPgwebInstallStep(null);
      setPgwebError(String(e));
    }
    setPgwebLoading(false);
  };

  const handleStartPgweb = async () => {
    setPgwebLoading(true);
    setPgwebError(null);
    try {
      await StartPgweb();
      await loadStatus();
    } catch (e: any) {
      console.error(e);
      setPgwebError(String(e));
    }
    setPgwebLoading(false);
  };

  const handleStopPgweb = async () => {
    setPgwebLoading(true);
    setPgwebError(null);
    try {
      await StopPgweb();
      await loadStatus();
    } catch (e: any) {
      console.error(e);
      setPgwebError(String(e));
    }
    setPgwebLoading(false);
  };

  const handleOpenPgweb = () => {
    OpenPgweb();
  };

  if (!statusLoaded) {
    return (
      <div>
        <div className="page-header">
          <h2>Database</h2>
          <p>Manage PostgreSQL databases</p>
        </div>
        <div className="card" style={{ color: "var(--text-muted)", fontSize: 13 }}>
          Loading...
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="page-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <h2>Database</h2>
          <p>Manage PostgreSQL databases</p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={loadStatus}>
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      {error && (
        <div className="card" style={{ marginBottom: 16, borderLeft: "3px solid var(--danger, #e53e3e)", color: "var(--text-secondary)", fontSize: 13 }}>
          {error}
        </div>
      )}

      {/* Not installed */}
      {!installed && (
        <div className="card empty-state">
          <Database size={32} />
          <h3>PostgreSQL is not installed</h3>
          <p>
            Install PostgreSQL 17 from pre-built binaries. This downloads the binary,
            extracts it, and initialises the data directory.
          </p>
          <button
            className="btn btn-primary"
            onClick={handleInstall}
            disabled={loading}
          >
            {installStep ?? "Install PostgreSQL"}
          </button>
        </div>
      )}

      {/* Installed but not running */}
      {installed && !running && (
        <div className="card" style={{ marginBottom: 16 }}>
          <div className="service-row">
            <div className="service-info">
              <span className="status-dot stopped" />
              <div>
                <div className="service-name">PostgreSQL</div>
                <div className="service-detail">Stopped</div>
              </div>
            </div>
            <button
              className="btn btn-primary btn-sm"
              onClick={handleStart}
              disabled={loading}
            >
              {loading ? "Starting..." : "Start"}
            </button>
          </div>
        </div>
      )}

      {/* Running */}
      {installed && running && (
        <>
          {/* Service controls */}
          <div className="card" style={{ marginBottom: 16 }}>
            <div className="service-row">
              <div className="service-info">
                <span className="status-dot running" />
                <div>
                  <div className="service-name">PostgreSQL</div>
                  <div className="service-detail">
                    Running &middot; Port 5432 &middot; Socket: postgres (trust auth)
                  </div>
                </div>
              </div>
              <button
                className="btn btn-danger btn-sm"
                onClick={handleStop}
                disabled={loading}
              >
                {loading ? "Stopping..." : "Stop"}
              </button>
            </div>
          </div>

          {/* Web Admin (pgweb) card — D-01, D-02, D-03 */}
          <div className="card" style={{ marginBottom: 16 }}>
            <div className="service-row">
              <div className="service-info">
                {pgwebInstalled && (
                  <span className={`status-dot ${pgwebRunning ? "running" : "stopped"}`} />
                )}
                <div>
                  <div className="service-name">Web Admin (pgweb)</div>
                  <div className="service-detail">
                    {!pgwebInstalled
                      ? "Not installed"
                      : pgwebRunning
                        ? "Running \u00B7 Port 8081"
                        : "Stopped"}
                  </div>
                </div>
              </div>
              <div className="service-actions">
                {!pgwebInstalled && (
                  <button
                    className="btn btn-primary btn-sm"
                    onClick={handleInstallPgweb}
                    disabled={pgwebLoading}
                  >
                    {pgwebInstallStep ?? "Install pgweb"}
                  </button>
                )}
                {pgwebInstalled && !pgwebRunning && (
                  <button
                    className="btn btn-primary btn-sm"
                    onClick={handleStartPgweb}
                    disabled={pgwebLoading}
                  >
                    {pgwebLoading ? "Starting..." : "Start"}
                  </button>
                )}
                {pgwebInstalled && pgwebRunning && (
                  <>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={handleStopPgweb}
                      disabled={pgwebLoading}
                    >
                      {pgwebLoading ? "Stopping..." : "Stop"}
                    </button>
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={handleOpenPgweb}
                    >
                      Open
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* pgweb error display */}
          {pgwebError && (
            <div className="card" style={{ marginBottom: 16, borderLeft: "3px solid var(--danger, #e53e3e)", color: "var(--text-secondary)", fontSize: 13 }}>
              {pgwebError}
            </div>
          )}

          {/* Create database */}
          <div className="card" style={{ marginBottom: 16 }}>
            <div className="card-title" style={{ marginBottom: 12 }}>Create Database</div>
            <div style={{ display: "flex", gap: 8 }}>
              <input
                type="text"
                className="input"
                placeholder="database_name"
                value={newDbName}
                onChange={(e) => setNewDbName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                style={{ flex: 1 }}
              />
              <button
                className="btn btn-primary btn-sm"
                onClick={handleCreate}
                disabled={creating || !newDbName.trim()}
              >
                <Plus size={14} /> {creating ? "Creating..." : "Create"}
              </button>
            </div>
          </div>

          {/* Database list */}
          <div className="card" style={{ marginBottom: 16 }}>
            <div className="card-title" style={{ marginBottom: 12 }}>Databases</div>
            {databases.length === 0 ? (
              <div style={{ color: "var(--text-muted)", fontSize: 13 }}>
                No user databases found
              </div>
            ) : (
              <table style={{ width: "100%", borderCollapse: "collapse" }}>
                <thead>
                  <tr>
                    <th style={{ textAlign: "left", fontSize: 12, color: "var(--text-muted)", paddingBottom: 8, fontWeight: 500 }}>
                      Name
                    </th>
                    <th style={{ textAlign: "right", fontSize: 12, color: "var(--text-muted)", paddingBottom: 8, fontWeight: 500 }}>
                      Size
                    </th>
                    <th style={{ width: 80 }} />
                  </tr>
                </thead>
                <tbody>
                  {databases.map((db) => (
                    <tr key={db.name} className="version-item" style={{ borderTop: "1px solid var(--border)" }}>
                      <td style={{ padding: "8px 0", fontWeight: 500 }}>{db.name}</td>
                      <td style={{ padding: "8px 0", textAlign: "right", color: "var(--text-muted)", fontSize: 12 }}>
                        {db.size}
                      </td>
                      <td style={{ textAlign: "right" }}>
                        <button className="btn btn-danger btn-sm" onClick={() => handleDrop(db.name)}>
                          <Trash2 size={13} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}

      {/* Drop confirmation modal */}
      {dropTarget && (
        <div style={{
          position: "fixed", inset: 0, background: "rgba(0,0,0,0.5)",
          display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
        }}>
          <div className="card" style={{ maxWidth: 400, width: "90%" }}>
            <h3 style={{ marginTop: 0, marginBottom: 8 }}>Drop Database</h3>
            <p style={{ color: "var(--text-secondary)", fontSize: 13, marginBottom: 16 }}>
              Are you sure you want to drop <strong>"{dropTarget}"</strong>? This action cannot be undone.
            </p>
            <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
              <button className="btn btn-secondary btn-sm" onClick={() => setDropTarget(null)} disabled={dropping}>
                Cancel
              </button>
              <button className="btn btn-danger btn-sm" onClick={confirmDrop} disabled={dropping}>
                {dropping ? "Dropping..." : "Drop Database"}
              </button>
            </div>
          </div>
        </div>
      )}

      {toastError && (
        <Toast error={toastError} onDismiss={() => setToastError(null)} />
      )}
    </div>
  );
}

export default PgDatabasePage;
