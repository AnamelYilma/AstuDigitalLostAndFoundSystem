// frontend/src/pages/Home.jsx
import styles from './Home.module.css';

export default function Home() {
  return (
    <div className={styles.container}>
      {/* HEADER */}
      <header className={styles.header}>
        <div className={styles.logo}>
          <div className={styles.logoIcon}>📦</div>
          <h1 className={styles.logoText}>ASTU Lost&Found</h1>
        </div>
        
        <div className={styles.headerActions}>
          <div className={styles.searchBar}>
            <input 
              type="text" 
              placeholder="Search..." 
              className={styles.searchInput}
            />
            <button className={styles.searchButton}>🔍</button>
          </div>
          
          <button className={styles.loginButton}>👤 Login</button>
          <button className={styles.registerButton}>📝 Register</button>
        </div>
      </header>

      {/* HERO SECTION */}
      <section className={styles.hero}>
        <h1 className={styles.heroTitle}>
          Lost something? <br />
          <span className={styles.heroHighlight}>We'll help you find it.</span>
        </h1>
        
        <p className={styles.heroDescription}>
          A centralized digital system for ASTU students to report, search, and track lost and found items securely.
        </p>
        
        <div className={styles.heroButtons}>
          <button className={styles.browseButton}>
            🔍 Browse Items
          </button>
          <button className={styles.reportButton}>
            ➕ Report Found Item
          </button>
        </div>
      </section>

      {/* FEATURES SECTION */}
      <section className={styles.features}>
        <div className={styles.featureCard}>
          <div className={styles.featureIcon}>🔍</div>
          <h2 className={styles.featureTitle}>Search & Filter</h2>
          <p className={styles.featureDesc}>
            Find your lost items with powerful search and filtering
          </p>
        </div>
        
        <div className={styles.featureCard}>
          <div className={styles.featureIcon}>🛡️</div>
          <h2 className={styles.featureTitle}>Secure Claims</h2>
          <p className={styles.featureDesc}>
            Verified claims process to keep your items safe
          </p>
        </div>
        
        <div className={styles.featureCard}>
          <div className={styles.featureIcon}>✏️</div>
          <h2 className={styles.featureTitle}>Easy Reporting</h2>
          <p className={styles.featureDesc}>
            Report lost or found items in seconds
          </p>
        </div>
      </section>
    </div>
  );
}