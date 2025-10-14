// src/components/RouteGuards.tsx
import React, { JSX } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const RequireAuth: React.FC<{ children: JSX.Element }> = ({ children }) => {
  const { isAuth, loading } = useAuth();
  const location = useLocation();

  if (loading) return <div>Загрузка...</div>;
  if (!isAuth) return <Navigate to="/signin" state={{ from: location }} replace />;
  return children;
};

export const RequireRole: React.FC<{ roles: string[]; children: JSX.Element }> = ({ roles, children }) => {
  const { user } = useAuth();
  if (!user) return <div>Нет доступа</div>;
  if (roles.includes(user.role)) return children;
  return <div>Доступ запрещён</div>; // можно заменить на Navigate на страницу "Нет доступа"
};
