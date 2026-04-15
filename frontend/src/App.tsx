import { useState } from "react";
import {
  LayoutDashboard,
  Globe,
  Server,
  FileCode2,
  Hexagon,
  ScrollText,
} from "lucide-react";
import Dashboard from "./pages/Dashboard";
import SitesPage from "./pages/SitesPage";
import ServicesPage from "./pages/ServicesPage";
import PhpPage from "./pages/PhpPage";
import NodePage from "./pages/NodePage";
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
            {navItems.map((item) => (
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
