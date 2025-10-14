import React, { useState, useEffect } from 'react';
import DefectCard from '../DefectCard/DefectCard';
import './DefectsList.scss';
import { getAttachmentsByDefect } from '../../../api/attachments';
import api from '../../../api/axios';

const API_ROOT = 'http://localhost:8080';

interface Defect {
  id: number;
  title: string;
  description: string;
  status: 'new' | 'in_progress' | 'review' | 'canceled';
  deadline: string;
  responsible_person_id: number;
  responsible_person_name?: string;
  image_url?: string;
}

interface DefectsListProps {
  buildingId: number | null;
  onDefectClick?: (defect: Defect) => void;
}

const DefectsList: React.FC<DefectsListProps> = ({ buildingId, onDefectClick }) => {
  const [defects, setDefects] = useState<Defect[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (buildingId) {
      fetchDefects(buildingId);
    } else {
      setDefects([]);
    }
  }, [buildingId]);

  const fetchDefects = async (buildingId: number) => {
    setLoading(true);
    setError('');
    
    try {
      const res = await api.get(`/defects?building_id=${buildingId}`);
      const data: Defect[] = res.data;

      // Если нет дефектов — быстро выйти
      if (!Array.isArray(data) || data.length === 0) {
        setDefects([]);
        setLoading(false);
        return;
      }

      // Подготовим массив промисов для получения первого вложения каждого дефекта
      const enhanced = await Promise.all(
        data.map(async (d) => {
          try {
            const atts = await getAttachmentsByDefect(d.id);
            if (Array.isArray(atts) && atts.length > 0) {
              const first = atts[0];
              // Преобразуем внутренний путь в публичный URL.
              // Примеры входных значений: "internal/uploads/defect_attachments/xxx.png"
              // Хотим: "http://localhost:8080/uploads/defect_attachments/xxx.png"
              let publicPath = first.url;

              // Если путь содержит "internal/", убираем его
              if (publicPath.startsWith('internal/')) {
                publicPath = publicPath.replace(/^internal\//, '');
              }

              // Убедимся, что путь начинается с uploads/
              if (!publicPath.startsWith('/')) {
                publicPath = `/${publicPath}`;
              }

              (d as Defect).image_url = `${API_ROOT}${publicPath}`;
            } else {
              (d as Defect).image_url = undefined;
            }
          } catch (err) {
            (d as Defect).image_url = undefined;
          }
          return d;
        })
      );

      setDefects(enhanced);
    } catch (error) {
      console.error('Ошибка загрузки дефектов:', error);
      setError('Ошибка загрузки дефектов');
    } finally {
      setLoading(false);
    }
  };

  if (!buildingId) {
    return (
      <div className="defectsList">
        <div className="defectsList__empty">
          <p>Выберите здание для просмотра дефектов</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="defectsList">
        <div className="defectsList__loading">
          <p>Загрузка дефектов...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="defectsList">
        <div className="defectsList__error">
          <p>{error}</p>
          <button 
            className="defectsList__retry-button"
            onClick={() => buildingId && fetchDefects(buildingId)}
          >
            Попробовать снова
          </button>
        </div>
      </div>
    );
  }

  if (defects.length === 0) {
    return (
      <div className="defectsList">
        <div className="defectsList__empty">
          <p>В выбранном здании нет дефектов</p>
        </div>
      </div>
    );
  }

  return (
    <div className="defectsList">
      <div className="defectsList__header">
        <h2 className="defectsList__title">Дефекты здания</h2>
        <span className="defectsList__count">дефектов: {defects.length}</span>
      </div>
      
      <div className="defectsList__container">
        {defects.map((defect) => (
          <DefectCard
            key={defect.id}
            defect={defect}
            onClick={onDefectClick}
          />
        ))}
      </div>
    </div>
  );
};

export default DefectsList;