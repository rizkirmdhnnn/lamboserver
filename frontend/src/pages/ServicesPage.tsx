import { useEffect, useState } from "react";
import { Play, Square, RefreshCw } from "lucide-react";
import {
  StartService,
  StopService,
  RestartService,
  GetAllStatuses,
} from "../../wailsjs/go/main/App";

function ServicesPage() {
  const [statuses, setStatuses] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState("");

  useEffect(() => {
    loadStatuses();
  }, []);

  const loadStatuses = async () => {
    try {
      const data = await GetAllStatuses();
      setStatuses(data || {});
    } catch (e) {
      console.error(e);
    }
  };

  const isRunning = (name: string) => statuses[name] === "running";

  const handleAction = async (action: string) => {
    setLoading(action);
    try {
      switch (action) {
        case "start-nginx": await StartService("nginx"); break;
        case "stop-nginx": await StopService("nginx"); break;
        case "reload-nginx": await RestartService("nginx"); break;
        case "start-dns": await StartService("dnsmasq"); break;
        case "stop-dns": await StopService("dnsmasq"); break;
      }
      await loadStatuses();
    } catch (e: any) {
      console.error(e);
    }
    setLoading("");
  };

  return (
    <div>
      <div className="page-header">
        <h2>Services</h2>
        <p>Manage your local development services</p>
      </div>

      <div className="card" style={{ marginBottom: 16 }}>
        <div className="service-row">
          <div className="service-info">
            <span className={`status-dot ${isRunning("nginx") ? "running" : "stopped"}`} />
            <div>
              <div className="service-name">Nginx</div>
              <div className="service-detail">
                {isRunning("nginx") ? "Running" : "Stopped"}
              </div>
            </div>
          </div>
          <div className="service-actions">
            {isRunning("nginx") ? (
              <>
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => handleAction("reload-nginx")}
                  disabled={loading === "reload-nginx"}
                >
                  <RefreshCw size={14} /> Reload
                </button>
                <button
                  className="btn btn-danger btn-sm"
                  onClick={() => handleAction("stop-nginx")}
                  disabled={loading === "stop-nginx"}
                >
                  <Square size={14} /> Stop
                </button>
              </>
            ) : (
              <button
                className="btn btn-primary btn-sm"
                onClick={() => handleAction("start-nginx")}
                disabled={loading === "start-nginx"}
              >
                <Play size={14} /> Start
              </button>
            )}
          </div>
        </div>

        <div className="service-row">
          <div className="service-info">
            <span className={`status-dot ${isRunning("dnsmasq") ? "running" : "stopped"}`} />
            <div>
              <div className="service-name">DNS (dnsmasq)</div>
              <div className="service-detail">
                {isRunning("dnsmasq") ? "Running" : "Stopped"}
              </div>
            </div>
          </div>
          <div className="service-actions">
            {isRunning("dnsmasq") ? (
              <button
                className="btn btn-danger btn-sm"
                onClick={() => handleAction("stop-dns")}
                disabled={loading === "stop-dns"}
              >
                <Square size={14} /> Stop
              </button>
            ) : (
              <button
                className="btn btn-primary btn-sm"
                onClick={() => handleAction("start-dns")}
                disabled={loading === "start-dns"}
              >
                <Play size={14} /> Start
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default ServicesPage;
