import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../api/axios';
import { getAttachmentsByDefect } from '../../api/attachments';
import './CreateDefectPage.scss';

type Priority = 'low' | 'medium' | 'high';

interface Building {
  id: number;
  name: string;
  address?: string;
  stage?: string;
}

interface User {
  id: number;
  login?: string;
  name?: string;
  lastname?: string;
  role?: string;
}

const API_ROOT = 'http://localhost:8080'; // использую для формирования публичных URL при необходимости

// helper: преобразует значение input[type=datetime-local] в "YYYY-MM-DD HH:mm:ss"
const formatDatetimeLocalToAPI = (value: string) => {
  // value как "2025-12-01T15:00"
  if (!value) return '';
  const d = new Date(value);
  const pad = (n: number) => n.toString().padStart(2, '0');
  const YYYY = d.getFullYear();
  const MM = pad(d.getMonth() + 1);
  const DD = pad(d.getDate());
  const hh = pad(d.getHours());
  const mm = pad(d.getMinutes());
  const ss = pad(d.getSeconds());
  return `${YYYY}-${MM}-${DD} ${hh}:${mm}:${ss}`;
};

const CreateDefectPage: React.FC = () => {
  const navigate = useNavigate();

  // form state
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [users, setUsers] = useState<User[] | null>(null); // если null — ещё не загружены; если [] — нет доступа/нет пользователей
  const [buildingId, setBuildingId] = useState<number | null>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<Priority>('medium');
  const [responsibleId, setResponsibleId] = useState<number | ''>('');
  const [deadlineLocal, setDeadlineLocal] = useState(''); // значение input[type=datetime-local]
  const [file, setFile] = useState<File | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    // fetch buildings and users
    fetchBuildings();
    fetchUsers();
  }, []);

  const fetchBuildings = async () => {
    try {
      const res = await api.get('/buildings');
      setBuildings(res.data || []);
      if (Array.isArray(res.data) && res.data.length > 0) {
        setBuildingId((prev) => prev ?? res.data[0].id);
      }
    } catch (err) {
      console.error('Ошибка загрузки зданий:', err);
      setBuildings([]);
    }
  };

  const fetchUsers = async () => {
    try {
      const res = await api.get('/users');
      setUsers(res.data || []);
    } catch (err) {
      // возможно нет доступа (403) — позволим вводить id вручную
      console.warn('Не удалось получить список пользователей (возможно нет прав):', err);
      setUsers([]); // ставим пустой массив => покажем поле для ручного ввода
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files && e.target.files[0];
    setFile(f ?? null);
  };

  const validate = () => {
    if (!buildingId) {
      setError('Выберите здание');
      return false;
    }
    if (!title.trim()) {
      setError('Введите название дефекта');
      return false;
    }
    if (!description.trim()) {
      setError('Введите описание дефекта');
      return false;
    }
    if (!responsibleId) {
      setError('Укажите ответственного (id)');
      return false;
    }
    // deadline optional? если обязателен - проверяем
    // if (!deadlineLocal) { setError('Укажите дедлайн'); return false; }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!validate()) return;

    setLoading(true);

    try {
      // форматируем дедлайн
      const deadline = deadlineLocal ? formatDatetimeLocalToAPI(deadlineLocal) : null;

      const payload: any = {
        building_id: buildingId,
        title: title.trim(),
        description: description.trim(),
        priority: priority,
        responsible_person_id: Number(responsibleId),
      };
      if (deadline) payload.deadline = deadline;
      // status optional, backend defaults to "new" — но можно явно задать:
      payload.status = 'new';

      // 1) Создаём дефект
      const res = await api.post('/defects', payload);
      const created: any = res.data;
      setSuccess('Дефект успешно создан');

      // 2) Если есть файл — загружаем вложение
      if (file && created && created.id) {
        const form = new FormData();
        form.append('file', file);
        // эндпоинт: POST /api/defects/:id/attachments
        await api.post(`/defects/${created.id}/attachments`, form, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });
        setSuccess('Дефект и вложение успешно загружены');
      }

      // redirect to defect page
      navigate(`/defects/${created.id}`);
    } catch (err: any) {
      console.error('Ошибка создания дефекта:', err);
      // нормализуем сообщение
      const msg = err?.response?.data?.error || err?.message || 'Ошибка создания дефекта';
      setError(typeof msg === 'string' ? msg : 'Ошибка создания дефекта');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-defect-page">
      <h2>Создание нового дефекта</h2>

      <form className="create-defect-form" onSubmit={handleSubmit}>
        <div className="form-row">
          <label>Здание</label>
          <select
            value={buildingId ?? ''}
            onChange={(e) => setBuildingId(Number(e.target.value))}
            required
          >
            <option value="" disabled>
              Выберите здание
            </option>
            {buildings.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name} {b.address ? `— ${b.address}` : ''}
              </option>
            ))}
          </select>
        </div>

        <div className="form-row">
          <label>Название дефекта</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Короткое название"
            required
          />
        </div>

        <div className="form-row">
          <label>Описание</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Подробное описание дефекта"
            rows={6}
            required
          />
        </div>

        <div className="form-row-grid">
          <div>
            <label>Приоритет</label>
            <select value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
              <option value="low">Низкий</option>
              <option value="medium">Средний</option>
              <option value="high">Высокий</option>
            </select>
          </div>

          <div>
            <label>Дедлайн</label>
            <input
              type="datetime-local"
              value={deadlineLocal}
              onChange={(e) => setDeadlineLocal(e.target.value)}
            />
          </div>
        </div>

        <div className="form-row">
          <label>Ответственный</label>

          {users && users.length > 0 ? (
            <select
              value={String(responsibleId)}
              onChange={(e) => setResponsibleId(Number(e.target.value))}
              required
            >
              <option value="">Выберите ответственного</option>
              {users.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.name || u.login} {u.lastname ? u.lastname : ''} {u.role ? `(${u.role})` : ''}
                </option>
              ))}
            </select>
          ) : (
            <input
              type="number"
              value={String(responsibleId)}
              onChange={(e) => setResponsibleId(e.target.value === '' ? '' : Number(e.target.value))}
              placeholder="Введите id ответственного (например 2)"
              required
            />
          )}
        </div>

        <div className="form-row">
          <label>Картинка (вложение)</label>
          <input className='file-upload-input' type="file" accept="image/*" onChange={handleFileChange} />
          <div className="hint">Изображение загружается после создания дефекта (если выбрано)</div>
        </div>

        {error && <div className="form-error">{error}</div>}
        {success && <div className="form-success">{success}</div>}

        <div className="form-actions">
          <button type="submit" disabled={loading}>
            {loading ? 'Сохраняем...' : 'Создать дефект'}
          </button>
          <button type="button" onClick={() => navigate(-1)} disabled={loading} className="btn-secondary">
            Отмена
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreateDefectPage;
