import StatusBadge from "./StatusBadge";

export default function ReportCard({ report }) {
  const { type, category, description, location, date, status, imageUrl } = report;

  return (
    <div className="report-card">
      <div className="report-card-image">
        {imageUrl
          ? <img src={imageUrl} alt={category} />
          : <span>📦</span>
        }
      </div>

      <div className="report-card-body">
        <div className="report-card-top">
          <span className={`report-type-tag ${type}`}>{type}</span>
          <StatusBadge status={status} />
        </div>

        <h3 className="report-card-title">{category}</h3>
        <p className="report-card-desc">{description}</p>

        <div className="report-card-meta">
          <span className="report-meta-item">📍 {location}</span>
          <span className="report-meta-item">📅 {date}</span>
        </div>

        <div className="report-card-footer">
          <button className="report-card-btn">View Details →</button>
        </div>
      </div>
    </div>
  );
}