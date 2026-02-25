export default function Navbar({ user }) {
  const initials = user?.name
    ?.split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2) ?? "U";

  return (
    <nav className="navbar">
      <div className="navbar-left">
        <div className="navbar-search">
          <span className="search-icon">🔍</span>
          <input type="text" placeholder="Search lost or found items…" />
        </div>
      </div>

      <div className="navbar-right">
        <div className="navbar-notification">
          <span className="notification-icon">🔔</span>
          <span className="notification-badge">2</span>
        </div>

        <div className="user-menu">
          <div className="user-avatar">{initials}</div>
          <div className="user-info">
            <span className="user-name">{user?.name ?? "Student"}</span>
            <span className="user-role">Student</span>
          </div>
          <span className="chevron-icon">▾</span>
        </div>
      </div>
    </nav>
  );
}