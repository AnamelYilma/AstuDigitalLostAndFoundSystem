import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Home from './pages/Home';
import Login from './pages/Login';
import Report from './pages/Report';
import AdminDashboard from './pages/AdminDashboard';
import ProtectedRoute from './utils/ProtectedRoute';
import './App.css';

const Navbar = () => {
    const { user, logout } = useAuth();

    return (
        <nav style={{ padding: '1rem', backgroundColor: '#f8f9fa', borderBottom: '1px solid #dee2e6', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
                <Link to="/" style={{ marginRight: '1rem', fontWeight: 'bold', textDecoration: 'none', color: '#333' }}>Lost & Found</Link>
                <Link to="/" style={{ marginRight: '1rem', textDecoration: 'none', color: '#555' }}>Home</Link>
                {user && <Link to="/report" style={{ marginRight: '1rem', textDecoration: 'none', color: '#555' }}>Report Item</Link>}
                {user?.role === 'admin' && <Link to="/admin" style={{ textDecoration: 'none', color: 'red' }}>Admin Panel</Link>}
            </div>
            <div>
                {user ? (
                    <>
                        <span style={{ marginRight: '1rem' }}>Welcome, <strong>{user.email}</strong></span>
                        <button onClick={logout} style={{ padding: '5px 10px', cursor: 'pointer' }}>Logout</button>
                    </>
                ) : (
                    <Link to="/login" style={{ textDecoration: 'none', color: '#007bff' }}>Login</Link>
                )}
            </div>
        </nav>
    );
};

function App() {
    return (
        <AuthProvider>
            <Router>
                <Navbar />
                <div className="container">
                    <Routes>
                        <Route path="/" element={<Home />} />
                        <Route path="/login" element={<Login />} />
                        <Route 
                            path="/report" 
                            element={
                                <ProtectedRoute>
                                    <Report />
                                </ProtectedRoute>
                            } 
                        />
                        <Route 
                            path="/admin" 
                            element={
                                <ProtectedRoute adminOnly={true}>
                                    <AdminDashboard />
                                </ProtectedRoute>
                            } 
                        />
                    </Routes>
                </div>
            </Router>
        </AuthProvider>
    );
}

export default App;
