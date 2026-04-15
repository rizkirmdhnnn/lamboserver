import { useEffect, useState } from "react";
import { Globe, Trash2, Plus, FolderOpen } from "lucide-react";
import { GetSites, LinkSite, UnlinkSite, SelectDirectory } from "../../wailsjs/go/main/App";

interface Site {
  domain: string;
  path: string;
  php_version: string;
  ssl_enabled: boolean;
  created_at: string;
}

function SitesPage() {
  const [sites, setSites] = useState<Site[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [newDomain, setNewDomain] = useState("");
  const [newPath, setNewPath] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadSites();
  }, []);

  const loadSites = async () => {
    try {
      const data = await GetSites();
      setSites(data || []);
    } catch (e) {
      console.error(e);
    }
  };

  const handleLink = async () => {
    if (!newPath || !newDomain) return;
    setLoading(true);
    try {
      await LinkSite(newPath, newDomain);
      await loadSites();
      setShowForm(false);
      setNewDomain("");
      setNewPath("");
    } catch (e) {
      console.error(e);
    }
    setLoading(false);
  };

  const handleUnlink = async (domain: string) => {
    try {
      await UnlinkSite(domain);
      await loadSites();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div>
      <div className="page-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div>
          <h2>Sites</h2>
          <p>Manage your linked project sites</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} /> Link Site
        </button>
      </div>

      {showForm && (
        <div className="card" style={{ marginBottom: 16 }}>
          <div className="card-title" style={{ marginBottom: 12 }}>Link New Site</div>
          <div style={{ display: "flex", gap: 12, marginBottom: 12 }}>
            <input
              type="text"
              placeholder="Domain (e.g. myapp)"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              style={{
                flex: 1,
                padding: "8px 12px",
                borderRadius: 6,
                border: "1px solid var(--border)",
                background: "var(--bg-primary)",
                color: "var(--text-primary)",
                fontSize: 13,
                outline: "none",
              }}
            />
            <div style={{ display: "flex", alignItems: "center", gap: 8, flex: 2 }}>
              <input
                type="text"
                placeholder="Project path (e.g. /Users/you/projects/myapp)"
                value={newPath}
                onChange={(e) => setNewPath(e.target.value)}
                style={{
                  flex: 1,
                  padding: "8px 12px",
                  borderRadius: 6,
                  border: "1px solid var(--border)",
                  background: "var(--bg-primary)",
                  color: "var(--text-primary)",
                  fontSize: 13,
                  outline: "none",
                }}
              />
              <FolderOpen
                size={18}
                style={{ color: "var(--text-muted)", cursor: "pointer" }}
                onClick={async () => {
                  try {
                    const dir = await SelectDirectory();
                    if (dir) {
                      setNewPath(dir);
                      if (!newDomain) {
                        const parts = dir.split("/");
                        setNewDomain(parts[parts.length - 1]);
                      }
                    }
                  } catch (e) {
                    console.error(e);
                  }
                }}
              />
            </div>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <button className="btn btn-primary btn-sm" onClick={handleLink} disabled={loading}>
              {loading ? "Linking..." : "Link"}
            </button>
            <button className="btn btn-secondary btn-sm" onClick={() => setShowForm(false)}>
              Cancel
            </button>
          </div>
        </div>
      )}

      {sites.length === 0 ? (
        <div className="card empty-state">
          <Globe />
          <h3>No sites linked</h3>
          <p>Click "Link Site" to connect a project directory to a .test domain</p>
        </div>
      ) : (
        sites.map((site) => (
          <div key={site.domain} className="site-item">
            <div>
              <div className="site-domain">{site.domain}</div>
              <div className="site-path">{site.path}</div>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
              {site.ssl_enabled && (
                <span className="version-source">SSL</span>
              )}
              <button
                className="btn btn-danger btn-sm"
                onClick={() => handleUnlink(site.domain)}
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        ))
      )}
    </div>
  );
}

export default SitesPage;
