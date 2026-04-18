import { useEffect, useState } from "react";
import { ScrollText } from "lucide-react";
import { GetLogFiles, ReadLog } from "../../wailsjs/go/main/App";

interface LogFile {
  name: string;
  path: string;
  size: number;
}

function LogsPage() {
  const [logFiles, setLogFiles] = useState<LogFile[]>([]);
  const [selectedLog, setSelectedLog] = useState("");
  const [lines, setLines] = useState<string[]>([]);

  useEffect(() => {
    loadLogFiles();
  }, []);

  const loadLogFiles = async () => {
    try {
      const data = await GetLogFiles();
      setLogFiles(data || []);
    } catch (e) {
      console.error(e);
    }
  };

  const handleSelectLog = async (path: string) => {
    setSelectedLog(path);
    try {
      const data = await ReadLog(path, 200);
      setLines(data || []);
    } catch (e) {
      console.error(e);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div>
      <div className="page-header">
        <h2>Logs</h2>
        <p>View nginx and service log files</p>
      </div>

      {logFiles.length === 0 ? (
        <div className="card empty-state">
          <ScrollText />
          <h3>No log files found</h3>
          <p>Log files will appear here once services are running</p>
        </div>
      ) : (
        <>
          <div style={{ display: "flex", gap: 8, marginBottom: 16, flexWrap: "wrap" }}>
            {logFiles.map((lf) => (
              <button
                key={lf.path}
                className={`btn ${selectedLog === lf.path ? "btn-primary" : "btn-secondary"} btn-sm`}
                onClick={() => handleSelectLog(lf.path)}
              >
                {lf.name} ({formatSize(lf.size)})
              </button>
            ))}
          </div>

          {selectedLog && (
            <div className="log-container">
              {lines.length === 0 ? (
                <div style={{ color: "var(--text-muted)" }}>Log file is empty</div>
              ) : (
                lines.map((line, i) => (
                  <div key={i} className="log-line">
                    {line}
                  </div>
                ))
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default LogsPage;
