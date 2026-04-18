import { useEffect, useState } from "react";
import { Database, Trash2, Plus, RefreshCw, Download, ExternalLink } from "lucide-react";
import {
  StartService,
  StopService,
  GetAllStatuses,
  InstallVersion,
  InitService,
  ListServiceDatabases,
  CreateServiceDatabase,
  DropServiceDatabase,
  InstallWebAdmin,
  UninstallWebAdmin,
  OpenWebAdmin,
} from "../../wailsjs/go/main/App";
import Toast, { ToastError } from "../components/Toast";

function DatabasePage() {
  const [installed, setInstalled] = useState(false);
  const [running, setRunning] = useState(false);
  const [databases, setDatabases] = useState<string[]>([]);
  const [newDbName, setNewDbName] = useState("");
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [installStep, setInstallStep] = useState<string | null>(null);
  const [toastError, setToastError] = useState<ToastError | null>(null);
  const [dropTarget, setDropTarget] = useState<string | null>(null);
  const [dropping, setDropping] = useState(false);
  const [statusLoaded, setStatusLoaded] = useState(false);
  const [pmaInstalled, setPmaInstalled] = useState(false);
  const [pmaLoading, setPmaLoading] = useState("");
  const [pmaError, setPmaError] = useState<string | null>(null);
  const [pmaConfirmUninstall, setPmaConfirmUninstall] = useState(false);

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      const statuses = await GetAllStatuses();
      const mysqlStatus = statuses["mysql"] || "not_installed";
      const isRunning = mysqlStatus === "running";
      setRunning(isRunning);
      setInstalled(mysqlStatus !== "not_installed");
      setPmaInstalled(statuses["phpmyadmin"] === "installed");
      setStatusLoaded(true);
      if (isRunning) {
        await loadDatabases();
      } else {
        setDatabases([]);
      }
    } catch (e) {
      console.error("Failed to load MySQL status:", e);
      setStatusLoaded(true);
    }
  };

  const loadDatabases = async () => {
    try {
      const dbs = await ListServiceDatabases("mysql");
      setDatabases((dbs || []).map((d: any) => d.name));
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
      await InstallVersion("mysql", "");
      setInstallStep("Initializing...");
      await InitService("mysql");
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
      await StartService("mysql");
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
      await StopService("mysql");
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
      await CreateServiceDatabase("mysql", newDbName.trim());
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
      await DropServiceDatabase("mysql", dropTarget);
      await loadDatabases();
    } catch (e: any) {
      console.error(e);
      setError(String(e));
    }
    setDropping(false);
    setDropTarget(null);
  };

  const handlePmaAction = async (action: string) => {
    setPmaLoading(action);
    setPmaError(null);
    try {
      switch (action) {
        case "install": await InstallWebAdmin("phpmyadmin", ""); break;
        case "uninstall": await UninstallWebAdmin("phpmyadmin"); break;
        case "open": await OpenWebAdmin("phpmyadmin"); break;
      }
      await loadStatus();
    } catch (e: any) {
      console.error(e);
      setPmaError(String(e));
    }
    setPmaLoading("");
  };

  if (!statusLoaded) {
    return (
      <div>
        <div className="page-header">
          <h2>Database</h2>
          <p>Manage MySQL databases</p>
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
          <p>Manage MySQL databases</p>
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

      {!installed && (
        <div className="card empty-state">
          <Database size={32} />
          <h3>MySQL is not installed</h3>
          <p>
            Install MySQL 8.4 LTS from dev.mysql.com. This downloads the binary,
            extracts it, and initialises the data directory.
          </p>
          <button
            className="btn btn-primary"
            onClick={handleInstall}
            disabled={loading}
          >
            {installStep ?? "Install MySQL"}
          </button>
        </div>
      )}

      {installed && !running && (
        <div className="card" style={{ marginBottom: 16 }}>
          <div className="service-row">
            <div className="service-info">
              <span className="status-dot stopped" />
              <div>
                <div className="service-name">MySQL</div>
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

      {installed && running && (
        <>
          <div className="card" style={{ marginBottom: 16 }}>
            <div className="service-row">
              <div className="service-info">
                <span className="status-dot running" />
                <div>
                  <div className="service-name">MySQL</div>
                  <div className="service-detail">
                    Running &middot; Port 3306 &middot; Socket: root (no password)
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
                    <th style={{ width: 80 }} />
                  </tr>
                </thead>
                <tbody>
                  {databases.map((db) => (
                    <tr key={db} className="version-item" style={{ borderTop: "1px solid var(--border)" }}>
                      <td style={{ padding: "8px 0", fontWeight: 500 }}>{db}</td>
                      <td style={{ textAlign: "right" }}>
                        <button
                          className="btn btn-danger btn-sm"
                          onClick={() => handleDrop(db)}
                        >
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

      {/* phpMyAdmin */}
      <div className="card" style={{ marginBottom: 16 }}>
        {pmaError && (
          <div style={{ padding: "12px 16px", borderLeft: "3px solid var(--danger)", color: "var(--text-secondary)", fontSize: 13 }}>
            {pmaError}
          </div>
        )}
        <div className="service-row">
          <div className="service-info">
            <span className={`status-dot ${pmaInstalled ? "running" : "stopped"}`} />
            <div>
              <div className="service-name">phpMyAdmin</div>
              <div className="service-detail">
                {pmaInstalled ? "Installed \u00b7 https://phpmyadmin.test" : "Not installed \u00b7 Web-based MySQL manager"}
              </div>
            </div>
          </div>
          <div className="service-actions">
            {pmaInstalled ? (
              <>
                {pmaConfirmUninstall ? (
                  <>
                    <span style={{ fontSize: 12, color: "var(--text-secondary)", marginRight: 8 }}>Are you sure?</span>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => {
                        setPmaConfirmUninstall(false);
                        handlePmaAction("uninstall");
                      }}
                      disabled={pmaLoading === "uninstall"}
                    >
                      <Trash2 size={14} /> {pmaLoading === "uninstall" ? "Uninstalling..." : "Yes, Uninstall"}
                    </button>
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => setPmaConfirmUninstall(false)}
                    >
                      Cancel
                    </button>
                  </>
                ) : (
                  <button
                    className="btn btn-danger btn-sm"
                    onClick={() => setPmaConfirmUninstall(true)}
                    disabled={pmaLoading === "uninstall"}
                  >
                    <Trash2 size={14} /> Uninstall
                  </button>
                )}
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => handlePmaAction("open")}
                >
                  <ExternalLink size={14} /> Open
                </button>
              </>
            ) : (
              <button
                className="btn btn-primary btn-sm"
                onClick={() => handlePmaAction("install")}
                disabled={pmaLoading === "install"}
              >
                <Download size={14} /> {pmaLoading === "install" ? "Installing..." : "Install phpMyAdmin"}
              </button>
            )}
          </div>
        </div>
      </div>

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

export default DatabasePage;
