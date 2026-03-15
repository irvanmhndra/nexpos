import axios from 'axios';
import { API_BASE_URL } from '@/utils/constants';

export const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  config => {
    // Get token from localStorage (matches AuthContext storage key)
    const token = localStorage.getItem('nexpos_access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      // Clear all auth keys (must match AuthContext storage keys exactly)
      localStorage.removeItem('nexpos_user');
      localStorage.removeItem('nexpos_access_token');
      localStorage.removeItem('nexpos_refresh_token');
      window.location.replace('/login');
    }
    return Promise.reject(error);
  }
);
