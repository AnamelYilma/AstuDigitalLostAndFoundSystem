// frontend/src/pages/Login.jsx
export default function Login({ onLogin }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    
    // SIMULATED LOGIN (replace with real API later)
    if (email && password) {
      const user = { 
        id: 1, 
        name: "Student", 
        role: email.includes('admin') ? 'admin' : 'student' 
      };
      localStorage.setItem('user', JSON.stringify(user));
      onLogin(user);
    } else {
      setError("Please enter email and password");
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.card}>
        <h1 style={styles.header}>ASTU Lost & Found</h1>
        
        {error && <div style={styles.error}>{error}</div>}
        
        <form onSubmit={handleSubmit} style={styles.form}>
          <input
            type="email"
            placeholder="Email (use admin@astu.edu for admin)"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            style={styles.input}
            required
          />
          
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            style={styles.input}
            required
          />
          
          <button type="submit" style={styles.button}>
            Login
          </button>
        </form>
        
        <div style={styles.register}>
          New user? <span style={styles.link}>Register here</span>
        </div>
      </div>
    </div>
  );
}

// SIMPLE STYLES (NO TAILWIND)
const styles = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    background: '#f5f5f5'
  },
  card: {
    background: 'white',
    padding: '30px',
    borderRadius: '10px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    width: '100%',
    maxWidth: '400px'
  },
  header: {
    textAlign: 'center',
    color: '#1a365d',
    marginBottom: '20px'
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
  button: {
    background: '#1a365d',
    color: 'white',
    border: 'none',
    padding: '12px',
    borderRadius: '4px',
    fontSize: '16px',
    cursor: 'pointer',
    marginTop: '10px'
  },
  error: {
    color: '#e53e3e',
    textAlign: 'center',
    marginBottom: '10px'
  },
  register: {
    textAlign: 'center',
    marginTop: '20px',
    color: '#4a5568'
  },
  link: {
    color: '#1a365d',
    fontWeight: 'bold',
    cursor: 'pointer'
  }
};