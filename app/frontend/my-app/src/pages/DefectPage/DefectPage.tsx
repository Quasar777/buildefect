import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import api from '../../api/axios';
import './DefectPage.scss';

interface Defect {
  id: number;
  building_id: number;
  building?: { id: number; name: string };
  title: string;
  description: string;
  status: 'new' | 'in_progress' | 'review' | 'closed' | 'canceled';
  deadline: string;
  created_by?: { id: number; name: string; lastname: string };
  responsible_person_id: number;
  responsible?: { id: number; name: string; lastname: string };
  updated_at: string;
}

interface Comment {
  id: number;
  user_id: number;
  user_name: string;
  user_lastname: string;
  text: string;
  created_at: string;
}

const DefectPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();

  const [defect, setDefect] = useState<Defect | null>(null);
  const [attachmentUrl, setAttachmentUrl] = useState<string | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState('');

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [editingResponsible, setEditingResponsible] = useState(false);
  const [allUsers, setAllUsers] = useState<{ id: number; name: string; lastname: string }[]>([]);
  const [selectedResponsibleId, setSelectedResponsibleId] = useState<number | null>(null);

  useEffect(() => {
    const fetchDefect = async () => {
      setLoading(true);
      try {
        // Получаем дефект
        const res = await api.get(`/defects/${id}`);
        const defectData: Defect = res.data;
        setDefect(defectData);

        // Получаем фото дефекта
        const attachRes = await api.get(`/defects/${id}/attachments`);
        if (attachRes.data?.length > 0) {
          const filePath = attachRes.data[0].url.replace(/^internal\//, '');
          setAttachmentUrl(`http://localhost:8080/${filePath}`);
        }

        // Получаем комментарии
        const commentsRes = await api.get(`/comments?defect_id=${id}`);
        const commentsData = commentsRes.data;

        // Получаем уникальные user_id
        const userIds = Array.from(new Set(commentsData.map((c: any) => c.created_by)));

        // Подгружаем пользователей параллельно
        const usersRes = await Promise.all(
          userIds.map((uid) => api.get(`/users/${uid}`))
        );
        const usersMap = new Map<number, { name: string; lastname: string }>();
        usersRes.forEach((r) => {
          const u = r.data;
          usersMap.set(u.id, { name: u.name, lastname: u.lastname });
        });

        // Обогащаем комментарии данными пользователя
        const commentsWithUser = commentsData.map((c: any) => ({
          ...c,
          user_name: usersMap.get(c.created_by)?.name || 'Unknown',
          user_lastname: usersMap.get(c.created_by)?.lastname || '',
        }));

        setComments(commentsWithUser);

        // Получаем всех пользователей для смены ответственного
        const usersListRes = await api.get('/users');
        setAllUsers(usersListRes.data);

      } catch (err: any) {
        setError(err?.response?.data?.error || 'Ошибка загрузки дефекта');
      } finally {
        setLoading(false);
      }
    };

    fetchDefect();
  }, [id]);


  const updateStatus = async (status: Defect['status']) => {
    if (!defect) return;
    try {
      await api.patch(`/defects/${defect.id}`, { status });
      setDefect({ ...defect, status, updated_at: new Date().toISOString() });
    } catch (err) {
      console.error(err);
    }
  };

  const deleteDefect = async () => {
    if (!defect) return;
    if (!window.confirm('Вы уверены, что хотите удалить дефект?')) return;
    try {
      await api.delete(`/defects/${defect.id}`);
      navigate('/');
    } catch (err) {
      console.error(err);
    }
  };

  const addComment = async () => {
  if (!newComment.trim() || !defect) return;
  try {
    const res = await api.post(`/comments`, { 
      text: newComment,
      defect_id: defect.id
    });

    // Обогащаем комментарий данными текущего пользователя
    const newCommentWithUser: Comment = {
      ...res.data,
      user_name: user?.name || 'Unknown',
      user_lastname: user?.lastname || ''
    };

    setComments([...comments, newCommentWithUser]);
    setNewComment('');
  } catch (err) {
    console.error(err);
  }
};


  const handleChangeResponsibleClick = () => {
    setEditingResponsible(true);
    setSelectedResponsibleId(defect?.responsible_person_id || null);
  };

  if (loading) return <p>Загрузка...</p>;
  if (error || !defect) return <p>{error || 'Дефект не найден'}</p>;

  const isEngineer = user?.role === 'engineer';
  const isManagerOrObserver = user?.role === 'manager' || user?.role === 'observer';

  return (
    <div className="defect-page">
      <div className="defect-page__back">
        <button onClick={() => navigate(-1)}>← Назад</button>
      </div>

      <h2>{defect.title} ({defect.building?.name || `Здание #${defect.building_id}`})</h2>

      <div className="defect-page__main">
        {attachmentUrl && (
          <div className="defect-page__image">
            <img src={attachmentUrl} alt={defect.title} />
          </div>
        )}

        <div className="defect-page__details">
          <p><b>Описание:</b> {defect.description}</p>
          <p><b>Статус:</b> {defect.status}</p>
          <p><b>Дедлайн:</b> {defect.deadline}</p>
          <p><b>Последнее</b> обновление: {defect.updated_at}</p>

          <p>
            <b>Ответственный:</b> {defect.responsible?.name} {defect.responsible?.lastname}
            {isManagerOrObserver && !editingResponsible && (
              <button onClick={handleChangeResponsibleClick}>Поменять</button>
            )}
          </p>

          {editingResponsible && (
            <div className="responsible-editor">
              <select 
                value={selectedResponsibleId || ''}
                onChange={(e) => setSelectedResponsibleId(Number(e.target.value))}
              >
                {allUsers.map(u => (
                  <option key={u.id} value={u.id}>{u.name} {u.lastname}</option>
                ))}
              </select>
              <button onClick={async () => {
                if (!selectedResponsibleId || !defect) return;
                try {
                  await api.patch(`/defects/${defect.id}`, { responsible_person_id: selectedResponsibleId });
                  setDefect({
                    ...defect,
                    responsible_person_id: selectedResponsibleId,
                    responsible: allUsers.find(u => u.id === selectedResponsibleId)
                  });
                  setEditingResponsible(false);
                } catch (err) {
                  console.error(err);
                }
              }}>Сохранить</button>
              <button onClick={() => setEditingResponsible(false)}>Отмена</button>
            </div>
          )}

          <p><b>Создал:</b> {defect.created_by?.name} {defect.created_by?.lastname}</p>

          <div className="defect-page__actions">
            {isEngineer && (
              <>
                {defect.status === 'new' && (
                  <button onClick={() => updateStatus('in_progress')}>Приступить к работе</button>
                )}
                {defect.status === 'in_progress' && (
                  <button onClick={() => updateStatus('review')}>Завершить задачу</button>
                )}
              </>
            )}

            {isManagerOrObserver && (
              <>
                {defect.status !== 'closed' && (
                  <button onClick={() => updateStatus('closed')}>Подтвердить выполнение работы</button>
                )}
                <button onClick={deleteDefect}>Удалить дефект</button>
              </>
            )}
          </div>
        </div>
      </div>

      <div className="defect-page__comments">
        <h3>Комментарии</h3>
        {comments.map((c) => (
          <div key={c.id} className="comment-card">
            <p><b>{c.user_name} {c.user_lastname}</b></p>
            <p>{new Date(c.created_at).toLocaleString()}</p>
            <p>{c.text}</p>
            {user?.role === 'observer' && (
              <button>Удалить комментарий</button>
            )}
          </div>
        ))}

        <div className="defect-page__add-comment">
          <textarea 
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder="Добавить комментарий"
          />
          <button onClick={addComment}>Отправить</button>
        </div>
      </div>
    </div>
  );
};

export default DefectPage;
