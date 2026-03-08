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
      // Handle unauthorized - redirect to login
      localStorage.removeItem('nexpos_auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
