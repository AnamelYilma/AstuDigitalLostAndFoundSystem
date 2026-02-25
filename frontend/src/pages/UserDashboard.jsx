import { useState } from "react";
import Navbar from "../components/Navbar";
import Sidebar from "../components/Sidebar";
import StatCard from "../components/StatCard";
import ReportCard from "../components/ReportCard";

// ── Mock data ──────────────────────────────────────────────
const mockUser = {
  name: "Abebe Girma",
  studentId: "ugr/1234/15",
  phone: "0912345678",
};

const mockStats = [
  { title: "My Lost Reports",  value: 3, icon: "😞", color: "red",    sub: "Items you reported lost"    },
  { title: "My Found Reports", value: 1, icon: "🎉", color: "green",  sub: "Items you reported found"   },
  { title: "Pending Claims",   value: 1, icon: "⏳", color: "yellow", sub: "Awaiting admin review"      },
  { title: "Matched Items",    value: 1, icon: "🔗", color: "blue",   sub: "Possible matches found"     },
];

const mockReports = [
  {
    id: 1,
    type: "lost",
    category: "Mobile Phone",
    description: "Black Samsung Galaxy A14, cracked screen protector, lost near Library Block B.",
    location: "Library Block B",
    date: "2025-01-18",
    status: "pending",
    imageUrl: null,
  },
  {
    id: 2,
    type: "lost",
    category: "Calculator",
    description: "Casio FX-991ES scientific calculator with name sticker on back.",
    location: "Main Hall Room 201",
    date: "2025-01-15",
    status: "matched",
    imageUrl: null,
  },
  {
    id: 3,
    type: "found",
    category: "ID Card",
    description: "Found a student ID card near the cafeteria entrance.",
    location: "Cafeteria",
    date: "2025-01-20",
    status: "approved",
    imageUrl: null,
  },
];
// ──────────────────────────────────────────────────────────

export default function UserDashboard() {
  const [activePage, setActivePage] = useState("dashboard");

  return (
    <div className="app-layout">
      <Sidebar activePage={activePage} onNavigate={setActivePage} />
      <Navbar user={mockUser} />

      <main className="main-content">

        {/* Page Header */}
        <div className="page-header">
          <h1>Welcome back, {mockUser.name.split(" ")[0]} 👋</h1>
          <p>Here's an overview of your lost & found activity.</p>
        </div>

        {/* Stats */}
        <div className="stats-grid">
          {mockStats.map((stat, i) => (
            <StatCard key={i} {...stat} />
          ))}
        </div>

        {/* Quick Actions */}
        <div className="section">
          <div className="section-header">
            <h2>Quick Actions</h2>
          </div>
          <div className="quick-actions">
            <button
              className="action-btn action-btn-primary"
              onClick={() => setActivePage("report-lost")}
            >
              <span className="action-btn-icon">😞</span>
              Report Lost Item
            </button>
            <button
              className="action-btn action-btn-secondary"
              onClick={() => setActivePage("report-found")}
            >
              <span className="action-btn-icon">🎉</span>
              Report Found Item
            </button>
            <button
              className="action-btn action-btn-secondary"
              onClick={() => setActivePage("search")}
            >
              <span className="action-btn-icon">🔎</span>
              Search Items
            </button>
          </div>
        </div>

        {/* Recent Reports */}
        <div className="section">
          <div className="section-header">
            <h2>My Recent Reports</h2>
            <a href="#" className="section-link" onClick={() => setActivePage("reports")}>
              View all →
            </a>
          </div>

          {mockReports.length > 0 ? (
            <div className="reports-grid">
              {mockReports.map((report) => (
                <ReportCard key={report.id} report={report} />
              ))}
            </div>
          ) : (
            <div className="empty-state">
              <div className="empty-state-icon">📭</div>
              <h3>No reports yet</h3>
              <p>You haven't submitted any lost or found reports.</p>
            </div>
          )}
        </div>

      </main>
    </div>
  );
}