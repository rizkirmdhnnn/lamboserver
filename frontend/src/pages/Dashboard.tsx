import { useEffect, useState } from "react";
import { Globe, FileCode2, Hexagon, Server, Bug, Database } from "lucide-react";
import { GetDashboardStatus, EnableDebug, DisableDebug } from "../../wailsjs/go/main/App";

interface DashboardData {
  nginx_status: { installed: boolean; running: boolean };
  dns_status: { installed: boolean; running: boolean; resolver: boolean };
  mysql_status: { installed: boolean; running: boolean; port: number };
  postgresql_status: { installed: boolean; running: boolean; port: number };
  php_versions: number;
  node_versions: number;
  sites_count: number;
  active_php: string;
  active_node: string;
  first_run: boolean;
  debug_mode: boolean;
  shell_integrated: boolean;
}

function Dashboard() {
  const [status, setStatus] = useState<DashboardData | null>(null);

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      const data = await GetDashboardStatus();
      setStatus(data);
    } catch (e) {
      console.error("Failed to load dashboard:", e);
    }
  };

  return (
    <div>
      <div className="page-header">
        <h2>Dashboard</h2>
        <p>Overview of your local development environment</p>
      </div>

      <div className="grid-3" style={{ marginBottom: 24 }}>
        <div className="card stat-card">
          <div className="stat-icon">
            <Globe />
          </div>
          <div>
            <div className="stat-value">{status?.sites_count ?? 0}</div>
            <div className="stat-label">Linked Sites</div>
          </div>
        </div>

        <div className="card stat-card">
          <div className="stat-icon">
            <FileCode2 />
          </div>
          <div>
            <div className="stat-value">{status?.php_versions ?? 0}</div>
            <div className="stat-label">PHP Versions</div>
          </div>
        </div>

        <div className="card stat-card">
          <div className="stat-icon">
            <Hexagon />
          </div>
          <div>
            <div className="stat-value">{status?.node_versions ?? 0}</div>
            <div className="stat-label">Node Versions</div>
          </div>
        </div>
      </div>

      <div className="grid-2" style={{ marginBottom: 16 }}>
        <div className="card">
          <div className="card-title" style={{ marginBottom: 16 }}>
            Services
          </div>
          <div className="service-row">
            <div className="service-info">
              <span
                className={`status-dot ${status?.nginx_status?.running ? "running" : "stopped"}`}
              />
              <div>
                <div className="service-name">Nginx</div>
                <div className="service-detail">Web Server</div>
              </div>
            </div>
            <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
              {status?.nginx_status?.running ? "Running" : "Stopped"}
            </span>
          </div>
          <div className="service-row">
            <div className="service-info">
              <span
                className={`status-dot ${status?.dns_status?.running ? "running" : "stopped"}`}
              />
              <div>
                <div className="service-name">DNS</div>
                <div className="service-detail">
                  .test domain resolution
                </div>
              </div>
            </div>
            <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
              {status?.dns_status?.running ? "Running" : "Stopped"}
            </span>
          </div>
          <div className="service-row">
            <div className="service-info">
              <span
                className={`status-dot ${
                  !status?.mysql_status?.installed
                    ? "stopped"
                    : status?.mysql_status?.running
                    ? "running"
                    : "stopped"
                }`}
              />
              <div>
                <div className="service-name">MySQL</div>
                <div className="service-detail">Database server</div>
              </div>
            </div>
            <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
              {!status?.mysql_status?.installed
                ? "Not Installed"
                : status?.mysql_status?.running
                ? "Running"
                : "Stopped"}
            </span>
          </div>
          <div className="service-row">
            <div className="service-info">
              <span
                className={`status-dot ${
                  !status?.postgresql_status?.installed
                    ? "stopped"
                    : status?.postgresql_status?.running
                    ? "running"
                    : "stopped"
                }`}
              />
              <div>
                <div className="service-name">PostgreSQL</div>
                <div className="service-detail">Database server</div>
              </div>
            </div>
            <span style={{ fontSize: 12, color: "var(--text-muted)" }}>
              {!status?.postgresql_status?.installed
                ? "Not Installed"
                : status?.postgresql_status?.running
                ? "Running"
                : "Stopped"}
            </span>
          </div>
        </div>

        <div className="card">
          <div className="card-title" style={{ marginBottom: 16 }}>
            Active Versions
          </div>
          <div className="service-row">
            <div className="service-info">
              <Server size={18} />
              <div>
                <div className="service-name">PHP</div>
                <div className="service-detail">Active version</div>
              </div>
            </div>
            <span className="version-badge">
              {status?.active_php || "None"}
            </span>
          </div>
          <div className="service-row">
            <div className="service-info">
              <Hexagon size={18} />
              <div>
                <div className="service-name">Node.js</div>
                <div className="service-detail">Active version</div>
              </div>
            </div>
            <span className="version-badge">
              {status?.active_node || "None"}
            </span>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <Bug size={18} />
            <span className="card-title">Debug Mode</span>
          </div>
          <button
            className={`btn btn-sm ${status?.debug_mode ? "btn-danger" : "btn-secondary"}`}
            onClick={async () => {
              try {
                if (status?.debug_mode) {
                  await DisableDebug();
                } else {
                  await EnableDebug();
                }
                await loadStatus();
              } catch (e) {
                console.error(e);
              }
            }}
          >
            {status?.debug_mode ? "Disable" : "Enable"}
          </button>
        </div>
        <div style={{ fontSize: 13, color: "var(--text-secondary)" }}>
          {status?.debug_mode ? (
            <span>
              <span className="status-dot running" style={{ marginRight: 8 }} />
              Debug active — all actions are being logged to{" "}
              <code style={{ fontSize: 11, color: "var(--accent)" }}>~/.lamboserver/logs/debug.log</code>
            </span>
          ) : (
            <span>
              <span className="status-dot stopped" style={{ marginRight: 8 }} />
              Enable to log all actions for troubleshooting
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
