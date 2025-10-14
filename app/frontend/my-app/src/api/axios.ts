// src/api/axios.ts
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 15000,
});

// Request interceptor: добавляем токен из localStorage (если есть)
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error && error.response) {
      const status = error.response.status;
      // Получаем путь запроса (например "/auth/login" или "/defects")
      const requestUrl: string = error.config?.url || '';

      if (status === 401) {
        // НЕ делать автоматический редирект для auth-эндпоинтов,
        // чтобы форма логина/регистрации могла корректно показать ошибку.
        const isAuthEndpoint =
          requestUrl.includes('/auth') || requestUrl.includes('/login') || requestUrl.includes('/register') || requestUrl.includes('/me');

        if (!isAuthEndpoint) {
          // Очистим локал сторедж и перенаправим на логин только для защищённых запросов
          localStorage.removeItem('auth_token');
          localStorage.removeItem('auth_user');
          // Можно сделать Navigate через react-router, но interceptor не имеет доступа — используем location
          window.location.href = '/signin';
        }
      }

      // Можно добавить обработку 403/500 здесь (например toast/notification)
    }
    return Promise.reject(error);
  }
);

export default api;
