const navItems = [
  { icon: "🏠", label: "Dashboard",     id: "dashboard" },
  { icon: "📋", label: "My Reports",    id: "reports",      badge: null },
  { icon: "🔎", label: "Search Items",  id: "search" },
  { icon: "📬", label: "My Claims",     id: "claims",       badge: 1 },
  { icon: "👤", label: "Profile",       id: "profile" },
];

export default function Sidebar({ activePage, onNavigate }) {
  return (
    <aside className="sidebar">
      {/* Brand */}
      <div className="sidebar-brand">
        <a href="#" className="brand-logo">
          <div className="brand-icon">🔍</div>
          <div className="brand-text">
            <span className="brand-title">Lost & Found</span>
            <span className="brand-subtitle">ASTU STEM CLUB</span>
          </div>
        </a>
      </div>

      {/* Navigation */}
      <nav className="sidebar-nav">
        <div className="nav-section-label">Menu</div>

        {navItems.map((item) => (
          <button
            key={item.id}
            className={`nav-item ${activePage === item.id ? "active" : ""}`}
            onClick={() => onNavigate(item.id)}
          >
            <span className="nav-icon">{item.icon}</span>
            <span className="nav-label">{item.label}</span>
            {item.badge && (
              <span className="nav-badge">{item.badge}</span>
            )}
          </button>
        ))}
      </nav>

      {/* Footer */}
      <div className="sidebar-footer">
        <button className="logout-btn">
          <span className="nav-icon">🚪</span>
          <span>Logout</span>
        </button>
      </div>
    </aside>
  );
}