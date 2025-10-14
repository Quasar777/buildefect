// src/api/buildings.ts
import api from './axios';

export const getBuildings = async () => {
  const { data } = await api.get('/buildings');
  return data;
};

export const getBuilding = async (id: number) => {
  const { data } = await api.get(`/buildings/${id}`);
  return data;
};

// и т.д. (create, patch, delete) - требуют авторизации, axios транслирует токен
