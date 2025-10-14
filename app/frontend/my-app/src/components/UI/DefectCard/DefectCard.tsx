import React from 'react';
import './DefectCard.scss';

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

interface DefectCardProps {
  defect: Defect;
  onClick?: (defect: Defect) => void;
}

const DefectCard: React.FC<DefectCardProps> = ({ defect, onClick }) => {
  const getStatusText = (status: string) => {
    const statusMap = {
      new: 'Новая',
      in_progress: 'В работе',
      review: 'На проверке',
      canceled: 'Отменена'
    };
    return statusMap[status as keyof typeof statusMap] || status;
  };

  const getStatusClass = (status: string) => {
    const statusClassMap = {
      new: 'defectCard__status--new',
      in_progress: 'defectCard__status--in-progress',
      review: 'defectCard__status--review',
      canceled: 'defectCard__status--canceled'
    };
    return statusClassMap[status as keyof typeof statusClassMap] || '';
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const isOverdue = (deadline: string) => {
    return new Date(deadline) < new Date();
  };

  return (
    <div 
      className="defectCard"
      onClick={() => onClick?.(defect)}
    >
      <div className="defectCard__image">
        {defect.image_url ? (
        <img
          src={defect.image_url}
          alt={defect.title}
          className="defectCard__image-img"
          onError={(e) => {
            const img = e.currentTarget as HTMLImageElement;
            img.onerror = null;
            // либо заменить на локальный плейсхолдер в public (например /placeholder-image.png)
            img.src = '/placeholder-image.png';
          }}
        />
      ) : (
        <div className="defectCard__image-placeholder">
          <span>📷</span>
        </div>
      )}
      </div>

      <div className="defectCard__content">
        <div className="defectCard__header">
          <h3 className="defectCard__title">{defect.title}</h3>
          <span className={`defectCard__status ${getStatusClass(defect.status)}`}>
            {getStatusText(defect.status)}
          </span>
        </div>

        <p className="defectCard__description">{defect.description}</p>

        <div className="defectCard__details">
          <div className="defectCard__detail">
            <span className="defectCard__detail-label">Дедлайн:</span>
            <span className={`defectCard__detail-value ${isOverdue(defect.deadline) ? 'defectCard__detail-value--overdue' : ''}`}>
              {formatDate(defect.deadline)}
            </span>
          </div>

          <div className="defectCard__detail">
            <span className="defectCard__detail-label">Ответственный:</span>
            <span className="defectCard__detail-value">
              {defect.responsible_person_name || `Инженер #${defect.responsible_person_id}`}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DefectCard;
