import { useEffect, useState } from "react";
import { FileCode2, Download, Check } from "lucide-react";
import {
  GetPhpVersions,
  GetAvailablePhpVersions,
  InstallPhp,
  SetActivePhp,
  UninstallPhp,
} from "../../wailsjs/go/main/App";
import VersionList from "../components/VersionList";

interface PhpVersion {
  version: string;
  path: string;
  active: boolean;
  source: string;
}

interface AvailablePhpVersion {
  version: string;
  series: string;
  internal_version: number;
  installed: boolean;
}

function PhpPage() {
  const [versions, setVersions] = useState<PhpVersion[]>([]);
  const [available, setAvailable] = useState<AvailablePhpVersion[]>([]);
  const [installing, setInstalling] = useState("");
  const [showAvailable, setShowAvailable] = useState(false);

  useEffect(() => {
    loadVersions();
  }, []);

  const loadVersions = async () => {
    try {
      const data = await GetPhpVersions();
      setVersions(data || []);
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
      const data = await GetAvailablePhpVersions();
      setAvailable(data || []);
      setShowAvailable(true);
    } catch (e) {
      console.error(e);
    }
  };

  const handleInstall = async (version: string) => {
    setInstalling(version);
    try {
      await InstallPhp(version);
      await loadVersions();
      const data = await GetAvailablePhpVersions();
      setAvailable(data || []);
    } catch (e) {
      console.error(e);
    }
    setInstalling("");
  };

  const handleSetActive = async (version: string) => {
    try {
      await SetActivePhp(version);
      await loadVersions();
    } catch (e) {
      console.error(e);
    }
  };

  const handleUninstall = async (version: string) => {
    try {
      await UninstallPhp(version);
      await loadVersions();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div>
      <div className="page-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <h2>PHP</h2>
          <p>Manage PHP versions installed on your system</p>
        </div>
        <button className="btn btn-primary" onClick={loadAvailable}>
          <Download size={16} /> {showAvailable ? "Hide" : "Install"}
        </button>
      </div>

      {showAvailable && (
        <div className="card" style={{ marginBottom: 16 }}>
          <div className="card-title" style={{ marginBottom: 12 }}>Available PHP Versions</div>
          {available.length === 0 ? (
            <div style={{ color: "var(--text-muted)", fontSize: 13 }}>Loading...</div>
          ) : (
            available.map((v) => (
              <div key={v.series} className="version-item">
                <div>
                  <span style={{ fontWeight: 600 }}>PHP {v.version}</span>
                  <span style={{ fontSize: 12, color: "var(--text-muted)", marginLeft: 8 }}>
                    ({v.series} series)
                  </span>
                </div>
                {v.installed ? (
                  <span className="version-badge">
                    <Check size={12} style={{ marginRight: 4 }} />
                    Installed
                  </span>
                ) : (
                  <button
                    className="btn btn-primary btn-sm"
                    onClick={() => handleInstall(v.version)}
                    disabled={installing === v.version}
                  >
                    {installing === v.version ? "Installing..." : "Install"}
                  </button>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {versions.length === 0 ? (
        <div className="card empty-state">
          <FileCode2 />
          <h3>No PHP versions detected</h3>
          <p>Click "Install" to download a PHP version, or install via Homebrew: brew install php@8.3</p>
        </div>
      ) : (
        <VersionList
          versions={versions}
          label="PHP"
          showSource
          canUninstall={(v) => v.source === "manual"}
          onSetActive={handleSetActive}
          onUninstall={handleUninstall}
        />
      )}
    </div>
  );
}

export default PhpPage;
