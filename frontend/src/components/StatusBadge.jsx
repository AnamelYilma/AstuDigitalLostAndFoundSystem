const statusIcons = {
  pending: "🕐",
  approved: "✅",
  rejected: "❌",
  matched: "🔗",
};

export default function StatusBadge({ status }) {
  return (
    <span className={`status-badge ${status}`}>
      <span className="status-dot" />
      {statusIcons[status] ?? ""} {status}
    </span>
  );
}