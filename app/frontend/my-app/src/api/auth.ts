// src/api/auth.ts
import api from './axios';

interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
}

export const login = async (login: string, password: string): Promise<LoginResponse> => {
  const { data } = await api.post('/auth/login', { login, password });
  return data;
};

export const register = async (login: string, password: string, name: string, lastname: string) => {
  const { data } = await api.post('/auth/register', { login, password, name, lastname });
  return data;
};

export const me = async () => {
  const { data } = await api.get('/me'); // важно: backend route /api/me
  return data;
};
