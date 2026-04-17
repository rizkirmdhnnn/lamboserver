import { useState } from "react";
import {
  LayoutDashboard,
  Globe,
  Server,
  FileCode2,
  Hexagon,
  Database,
  ScrollText,
  ChevronDown,
} from "lucide-react";
import Dashboard from "./pages/Dashboard";
import SitesPage from "./pages/SitesPage";
import ServicesPage from "./pages/ServicesPage";
import PhpPage from "./pages/PhpPage";
import NodePage from "./pages/NodePage";
import DatabasePage from "./pages/DatabasePage";
import PgDatabasePage from "./pages/PgDatabasePage";
import LogsPage from "./pages/LogsPage";

const navItems = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
  { id: "sites", label: "Sites", icon: Globe },
  { id: "services", label: "Services", icon: Server },
  { id: "php", label: "PHP", icon: FileCode2 },
  { id: "node", label: "Node.js", icon: Hexagon },
  { id: "logs", label: "Logs", icon: ScrollText },
];

function App() {
  const [activePage, setActivePage] = useState("dashboard");
  const [dbOpen, setDbOpen] = useState(false);

  const renderPage = () => {
    switch (activePage) {
      case "dashboard":
        return <Dashboard />;
      case "sites":
        return <SitesPage />;
      case "services":
        return <ServicesPage />;
      case "php":
        return <PhpPage />;
      case "node":
        return <NodePage />;
      case "database":
        return <DatabasePage />;
      case "pg-database":
        return <PgDatabasePage />;
      case "logs":
        return <LogsPage />;
      default:
        return <Dashboard />;
    }
  };

  return (
    <>
      <div className="titlebar" />
      <div className="app-layout">
        <aside className="sidebar">
          <div className="sidebar-header">
            <h1>LamboServer</h1>
            <span>Local Dev Environment</span>
          </div>
          <nav className="sidebar-nav">
            {navItems.filter(item => ["dashboard", "sites", "services", "php", "node"].includes(item.id)).map((item) => (
              <button
                key={item.id}
                className={`nav-item ${activePage === item.id ? "active" : ""}`}
                onClick={() => setActivePage(item.id)}
              >
                <item.icon />
                {item.label}
              </button>
            ))}

            {/* Database dropdown — parent toggles only, does not navigate (per D-02) */}
            <div>
              <button
                className={`nav-item ${(activePage === "database" || activePage === "pg-database") ? "active" : ""}`}
                onClick={() => setDbOpen(!dbOpen)}
              >
                <Database />
                Database
                <ChevronDown
                  size={14}
                  style={{
                    marginLeft: "auto",
                    transform: dbOpen ? "rotate(180deg)" : "rotate(0deg)",
                    transition: "transform 0.15s",
                  }}
                />
              </button>
              {dbOpen && (
                <div style={{ paddingLeft: 16 }}>
                  <button
                    className={`nav-item ${activePage === "database" ? "active" : ""}`}
                    onClick={() => setActivePage("database")}
                  >
                    MySQL
                  </button>
                  <button
                    className={`nav-item ${activePage === "pg-database" ? "active" : ""}`}
                    onClick={() => setActivePage("pg-database")}
                  >
                    PostgreSQL
                  </button>
                </div>
              )}
            </div>

            {navItems.filter(item => item.id === "logs").map((item) => (
              <button
                key={item.id}
                className={`nav-item ${activePage === item.id ? "active" : ""}`}
                onClick={() => setActivePage(item.id)}
              >
                <item.icon />
                {item.label}
              </button>
            ))}
          </nav>
        </aside>
        <main className="main-content">{renderPage()}</main>
      </div>
    </>
  );
}

export default App;
