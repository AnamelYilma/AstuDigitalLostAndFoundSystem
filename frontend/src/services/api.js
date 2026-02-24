// frontend/src/services/api.js
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080', // Don't include /api here if your routes have it
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Don't set Content-Type for FormData
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type'];
    }
    
    // Log in development
    if (import.meta.env.DEV) {
      console.log('API Request:', config.method.toUpperCase(), config.url);
    }
    
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle common errors
    if (error.response) {
      switch (error.response.status) {
        case 401:
          console.log('Session expired - redirecting to login');
          // You might want to redirect to login here
          break;
        case 403:
          console.log('Access forbidden');
          break;
        case 500:
          console.log('Server error');
          break;
      }
    }
    return Promise.reject(error);
  }
);

export default api;