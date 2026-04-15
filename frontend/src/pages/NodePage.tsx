import { useEffect, useState } from "react";
import { Hexagon, Download } from "lucide-react";
import {
  GetNodeVersions,
  GetAvailableNodeVersions,
  InstallNode,
  SetActiveNode,
  UninstallNode,
} from "../../wailsjs/go/main/App";
import VersionList from "../components/VersionList";

interface NodeVersion {
  version: string;
  path: string;
  active: boolean;
}

function NodePage() {
  const [installed, setInstalled] = useState<NodeVersion[]>([]);
  const [available, setAvailable] = useState<string[]>([]);
  const [installing, setInstalling] = useState("");
  const [showAvailable, setShowAvailable] = useState(false);

  useEffect(() => {
    loadInstalled();
  }, []);

  const loadInstalled = async () => {
    try {
      const data = await GetNodeVersions();
      setInstalled(data || []);
    } catch (e) {
      console.error(e);
    }
  };

  const loadAvailable = async () => {
    if (available.length > 0) {
      setShowAvailable(!showAvailable);
      return;
    }
    try {
      const data = await GetAvailableNodeVersions();
      setAvailable(data || []);
      setShowAvailable(true);
    } catch (e) {
      console.error(e);
    }
  };

  const handleInstall = async (version: string) => {
    setInstalling(version);
    try {
      await InstallNode(version);
      await loadInstalled();
    } catch (e) {
      console.error(e);
    }
    setInstalling("");
  };

  const handleSetActive = async (version: string) => {
    try {
      await SetActiveNode(version);
      await loadInstalled();
    } catch (e) {
      console.error(e);
    }
  };

  const handleUninstall = async (version: string) => {
    try {
      await UninstallNode(version);
      await loadInstalled();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div>
      <div className="page-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <h2>Node.js</h2>
          <p>Manage Node.js versions</p>
        </div>
        <button className="btn btn-primary" onClick={loadAvailable}>
          <Download size={16} /> {showAvailable ? "Hide" : "Install"}
        </button>
      </div>

      {showAvailable && (
        <div className="card" style={{ marginBottom: 16 }}>
          <div className="card-title" style={{ marginBottom: 12 }}>Available LTS Versions</div>
          {available.map((v) => (
            <div key={v} className="version-item">
              <span style={{ fontWeight: 600 }}>{v}</span>
              <button
                className="btn btn-primary btn-sm"
                onClick={() => handleInstall(v)}
                disabled={installing === v}
              >
                {installing === v ? "Installing..." : "Install"}
              </button>
            </div>
          ))}
        </div>
      )}

      {installed.length === 0 ? (
        <div className="card empty-state">
          <Hexagon />
          <h3>No Node.js versions installed</h3>
          <p>Click "Install" to download a Node.js version</p>
        </div>
      ) : (
        <VersionList
          versions={installed}
          label="Node"
          canUninstall={() => true}
          onSetActive={handleSetActive}
          onUninstall={handleUninstall}
        />
      )}
    </div>
  );
}

export default NodePage;
