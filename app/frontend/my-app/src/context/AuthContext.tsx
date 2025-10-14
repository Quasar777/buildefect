import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import * as authApi from '../api/auth'; // <-- новый модуль

export interface User {
  id: number;
  login: string;
  name: string;
  lastname: string;
  role: 'engineer' | 'manager' | 'observer';
}

export interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuth: boolean;
  login: (login: string, password: string) => Promise<void>;
  register: (login: string, password: string, name: string, lastname: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

export const AuthContext = createContext<AuthContextType | null>(null);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const API_BASE_URL = 'http://localhost:8080/api';

  useEffect(() => {
    const savedToken = localStorage.getItem('auth_token');
    const savedUser = localStorage.getItem('auth_user');
    
    if (savedToken && savedUser) {
      setToken(savedToken);
      setUser(JSON.parse(savedUser));
    }
    setLoading(false);
  }, []);

  const loginUser = async (login: string, password: string) => {
    try {
      const data = await authApi.login(login, password);
      const accessToken = data.access_token;
      localStorage.setItem('auth_token', accessToken);
      setToken(accessToken);

      const meData = await authApi.me();
      setUser(meData);
      localStorage.setItem('auth_user', JSON.stringify(meData));
    } catch (err: any) {
      // Берем оригинальное сообщение
      let msg = err?.response?.data?.error || err.message || 'Ошибка авторизации';

      // ✅ Преобразуем системные ошибки в понятные пользователю
      if (msg.toLowerCase().includes('invalid credentials')) {
        msg = 'Неверный логин или пароль';
      } else if (msg.toLowerCase().includes('unauthorized')) {
        msg = 'Ошибка авторизации. Проверьте данные для входа';
      } else if (msg.toLowerCase().includes('user not found')) {
        msg = 'Пользователь не найден';
      }

      throw new Error(msg);
    }
  };

  const registerUser = async (login: string, password: string, name: string, lastname: string) => {
    try {
      await authApi.register(login, password, name, lastname);
      // После регистрации — автоматически залогиниваем
      await loginUser(login, password);
    } catch (err: any) {
      const msg = err?.response?.data?.error || err.message || 'Ошибка регистрации';
      throw new Error(msg);
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
    // редирект на страницу входа
    window.location.href = '/signin';
  };

  const value: AuthContextType = {
    user,
    token,
    isAuth: !!user && !!token,
    login: loginUser,
    register: registerUser,
    logout,
    loading,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
