// frontend/src/pages/Dashboard.jsx
import { useState } from 'react';

export default function Dashboard({ user, onLogout }) {
  const [activeTab, setActiveTab] = useState('lost');
  
  // SAMPLE ITEMS (replace with real data later)
  const items = [
    { id: 1, title: "Student ID Card", location: "Library", date: "2024-02-20", status: "PENDING" },
    { id: 2, title: "Blue Calculator", location: "Engineering Building", date: "2024-02-22", status: "CLAIMABLE" }
  ];

  return (
    <div style={styles.container}>
      {/* HEADER */}
      <header style={styles.header}>
        <div>
          <h1>ASTU Lost & Found System</h1>
          <p>Welcome, {user.name} ({user.role})</p>
        </div>
        <button onClick={onLogout} style={styles.logoutButton}>
          Logout
        </button>
      </header>

      {/* TABS */}
      <div style={styles.tabs}>
        <button 
          onClick={() => setActiveTab('lost')}
          style={activeTab === 'lost' ? styles.activeTab : styles.tab}
        >
          Report Lost Item
        </button>
        <button 
          onClick={() => setActiveTab('found')}
          style={activeTab === 'found' ? styles.activeTab : styles.tab}
        >
          Report Found Item
        </button>
        <button 
          onClick={() => setActiveTab('search')}
          style={activeTab === 'search' ? styles.activeTab : styles.tab}
        >
          Search Items
        </button>
      </div>

      {/* CONTENT */}
      <div style={styles.content}>
        {activeTab === 'lost' && <LostForm />}
        {activeTab === 'found' && <FoundForm />}
        {activeTab === 'search' && <SearchResults items={items} />}
      </div>
    </div>
  );
}

// FORM COMPONENTS
function LostForm() {
  return (
    <div style={styles.formCard}>
      <h2>Report Lost Item</h2>
      <form style={styles.form}>
        <input placeholder="Item name" style={styles.input} />
        <input placeholder="Location lost" style={styles.input} />
        <textarea placeholder="Description" style={{...styles.input, height: '100px'}} />
        <input type="file" style={styles.input} />
        <button type="submit" style={styles.submitButton}>
          Submit Report
        </button>
      </form>
    </div>
  );
}

function FoundForm() {
  return (
    <div style={styles.formCard}>
      <h2>Report Found Item</h2>
      <form style={styles.form}>
        <input placeholder="Item name" style={styles.input} />
        <input placeholder="Location found" style={styles.input} />
        <textarea placeholder="Description" style={{...styles.input, height: '100px'}} />
        <input type="file" style={styles.input} />
        <button type="submit" style={styles.submitButton}>
          Submit Report
        </button>
      </form>
    </div>
  );
}

function SearchResults({ items }) {
  return (
    <div>
      <div style={styles.searchBar}>
        <input placeholder="Search items..." style={styles.searchInput} />
        <select style={styles.filterSelect}>
          <option>All Categories</option>
          <option>ID Cards</option>
          <option>Electronics</option>
        </select>
      </div>
      
      <div style={styles.itemsGrid}>
        {items.map(item => (
          <div key={item.id} style={styles.itemCard}>
            <h3>{item.title}</h3>
            <p><strong>Location:</strong> {item.location}</p>
            <p><strong>Date:</strong> {item.date}</p>
            <p><strong>Status:</strong> 
              <span style={{
                ...styles.status,
                ...(item.status === 'PENDING' && { background: '#fed7d7', color: '#c53030' }),
                ...(item.status === 'CLAIMABLE' && { background: '#bee3f8', color: '#2b6cb0' })
              }}>
                {item.status}
              </span>
            </p>
            <button style={styles.claimButton}>Claim Item</button>
          </div>
        ))}
      </div>
    </div>
  );
}

// ALL STYLES IN ONE PLACE
const styles = {
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '20px'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '20px 0',
    marginBottom: '30px',
    borderBottom: '1px solid #eee'
  },
  logoutButton: {
    background: '#e53e3e',
    color: 'white',
    border: 'none',
    padding: '8px 16px',
    borderRadius: '4px',
    cursor: 'pointer'
  },
  tabs: {
    display: 'flex',
    gap: '10px',
    marginBottom: '20px'
  },
  tab: {
    padding: '10px 20px',
    background: '#f0f0f0',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer'
  },
  activeTab: {
    padding: '10px 20px',
    background: '#1a365d',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer'
  },
  content: {
    background: 'white',
    padding: '20px',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
  },
  formCard: {
    padding: '20px'
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '15px'
  },
  input: {
    padding: '12px',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '16px'
  },
  submitButton: {
    background: '#1a365d',
    color: 'white',
    border: 'none',
    padding: '12px',
    borderRadius: '4px',
    fontSize: '16px',
    cursor: 'pointer',
    marginTop: '10px'
  },
  searchBar: {
    display: 'flex',
    gap: '10px',
    marginBottom: '20px',
    flexWrap: 'wrap'
  },
  searchInput: {
    flex: 1,
    minWidth: '200px',
    padding: '10px',
    border: '1px solid #ddd',
    borderRadius: '4px'
  },
  filterSelect: {
    padding: '10px',
    border: '1px solid #ddd',
    borderRadius: '4px'
  },
  itemsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
    gap: '20px'
  },
  itemCard: {
    border: '1px solid #eee',
    borderRadius: '8px',
    padding: '15px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
  },
  status: {
    padding: '3px 8px',
    borderRadius: '4px',
    fontSize: '12px',
    fontWeight: 'bold',
    marginLeft: '8px'
  },
  claimButton: {
    background: '#38a169',
    color: 'white',
    border: 'none',
    padding: '8px 16px',
    borderRadius: '4px',
    marginTop: '10px',
    cursor: 'pointer'
  }
};