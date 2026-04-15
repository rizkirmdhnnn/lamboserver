import { useEffect, useState } from "react";
import { Play, Square, RefreshCw } from "lucide-react";
import {
  GetNginxStatus,
  GetDnsStatus,
  StartNginx,
  StopNginx,
  ReloadNginx,
  StartDns,
  StopDns,
} from "../../wailsjs/go/main/App";

function ServicesPage() {
  const [nginxStatus, setNginxStatus] = useState({ installed: false, running: false });
  const [dnsStatus, setDnsStatus] = useState({ installed: false, running: false, resolver: false });
  const [loading, setLoading] = useState("");

  useEffect(() => {
    loadStatuses();
  }, []);

  const loadStatuses = async () => {
    try {
      const [nginx, dns] = await Promise.all([GetNginxStatus(), GetDnsStatus()]);
      setNginxStatus(nginx);
      setDnsStatus(dns);
    } catch (e) {
      console.error(e);
    }
  };

  const handleAction = async (action: string) => {
    setLoading(action);
    try {
      switch (action) {
        case "start-nginx": await StartNginx(); break;
        case "stop-nginx": await StopNginx(); break;
        case "reload-nginx": await ReloadNginx(); break;
        case "start-dns": await StartDns(); break;
        case "stop-dns": await StopDns(); break;
      }
      await loadStatuses();
    } catch (e) {
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
            <span className={`status-dot ${nginxStatus.running ? "running" : "stopped"}`} />
            <div>
              <div className="service-name">Nginx</div>
              <div className="service-detail">
                {nginxStatus.running ? "Running" : "Stopped"}
              </div>
            </div>
          </div>
          <div className="service-actions">
            {nginxStatus.running ? (
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
            <span className={`status-dot ${dnsStatus.running ? "running" : "stopped"}`} />
            <div>
              <div className="service-name">DNS (dnsmasq)</div>
              <div className="service-detail">
                {dnsStatus.running ? "Running" : "Stopped"} &middot;{" "}
                Resolver: {dnsStatus.resolver ? "OK" : "Not configured"}
              </div>
            </div>
          </div>
          <div className="service-actions">
            {dnsStatus.running ? (
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
