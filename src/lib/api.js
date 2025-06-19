import axios from 'axios';

const API_HOST = process.env.REACT_APP_API_HOST || 'https://api.securesystem.email';

const instance = axios.create({
  baseURL: API_HOST,
  headers: { 'Content-Type': 'application/json' },
});

export const login = (data) => instance.post('/api/auth/login', data);
export const signup = (data) => instance.post('/api/auth/signup', data);
export const verifyTotp = (data) => instance.post('/api/auth/verify-totp', data); 