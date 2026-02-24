// frontend/src/context/AuthContext.jsx
import { createContext, useState, useContext, useEffect } from 'react';
import api from '../services/api'; // Your configured axios instance

const AuthContext = createContext();

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Check auth status with backend
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Backend checks HTTP-only cookie
        const response = await api.get('/auth/me');
        setUser(response.data.user);
      } catch (error) {
        setUser(null);
      } finally {
        setLoading(false);
      }
    };
    
    checkAuth();
  }, []);

  // Login - credentials sent to backend
  const login = async (email, password) => {
    try {
      setLoading(true);
      
      // Validate input
      if (!email || !password) {
        throw new Error('Email and password required');
      }
      
      // Backend handles:
      // - Password verification
      // - JWT generation
      // - Setting HTTP-only cookie
      const response = await api.post('/auth/login', { 
        email, 
        password 
      });
      
      setUser(response.data.user);
      return { success: true };
      
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Login failed' 
      };
    } finally {
      setLoading(false);
    }
  };

  // Logout
  const logout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
    }
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);