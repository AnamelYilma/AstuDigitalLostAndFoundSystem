// frontend/src/App.jsx
import { Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import Login from './pages/Login';
import Report from './pages/Report';
import ProtectedRoute from './utils/ProtectedRoute';

function App() {
  return (
    <Routes>
      {/* Public Routes */}
      <Route path="/" element={<Home />} />
      <Route path="/login" element={<Login />} />
      
      {/* Protected Routes */}
      <Route element={<ProtectedRoute />}>
        <Route path="/report" element={<Report />} />
      </Route>
      
      {/* Fallback */}
      <Route path="*" element={<Home />} />
    </Routes>
  );
}

export default App;