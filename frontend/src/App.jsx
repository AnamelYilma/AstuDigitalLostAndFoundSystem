// frontend/src/App.jsx
import { useState, useEffect } from 'react';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';

function App() {
  const [user, setUser] = useState(null);

  // Check if user is logged in (simplified)
  useEffect(() => {
    const savedUser = localStorage.getItem('user');
    if (savedUser) setUser(JSON.parse(savedUser));
  }, []);

  if (!user) return <Login onLogin={setUser} />;

  return <Dashboard user={user} onLogout={() => setUser(null)} />;
}

export default App;