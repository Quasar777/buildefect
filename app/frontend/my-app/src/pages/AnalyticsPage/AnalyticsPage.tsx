import React, { useEffect, useMemo, useState } from 'react';
import api from '../../api/axios';
import './AnalyticsPage.scss';
import { useParams, useNavigate } from 'react-router-dom';

import { saveAs } from 'file-saver';
import * as XLSX from 'xlsx';

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Tooltip,
  Legend,
  Title,
} from 'chart.js';
import { Doughnut, Bar } from 'react-chartjs-2';

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, Tooltip, Legend, Title);

type Defect = {
  id: number;
  title?: string;
  description?: string;
  status?: string;
  deadline?: string | null;
  building_id?: number;
  priority?: string;
  responsible_person_id?: number;
  created_at?: string;
  updated_at?: string;
};

type Building = {
  id: number;
  name: string;
  address?: string;
};

const parseDeadline = (val?: string | null): Date | null => {
  if (!val) return null;
  const s = val.includes('T') ? val : val.replace(' ', 'T');
  const d = new Date(s);
  if (isNaN(d.getTime())) return null;
  return d;
};

// === EXPORT SECTION START ===
const csvEscape = (value: any) => {
  if (value === null || value === undefined) return '';
  const s = String(value);
  if (/[",\n\r]/.test(s)) return `"${s.replace(/"/g, '""')}"`;
  return s;
};
// === EXPORT SECTION END ===

const AnalyticsPage: React.FC = () => {
  const [defects, setDefects] = useState<Defect[]>([]);
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchAll = async () => {
      setLoading(true);
      setError(null);
      try {
        const dRes = await api.get('/defects');
        const bRes = await api.get('/buildings');
        setDefects(Array.isArray(dRes.data) ? dRes.data : []);
        setBuildings(Array.isArray(bRes.data) ? bRes.data : []);
      } catch (err: any) {
        setError(err?.response?.data?.error || 'Ошибка загрузки данных аналитики');
      } finally {
        setLoading(false);
      }
    };
    fetchAll();
  }, []);

  const metrics = useMemo(() => {
    const now = new Date();
    const total = defects.length;

    const countByStatus = defects.reduce<Record<string, number>>((acc, d) => {
      const s = d.status || 'unknown';
      acc[s] = (acc[s] || 0) + 1;
      return acc;
    }, {});

    const closedCount = (countByStatus['closed'] || 0) + (countByStatus['canceled'] || 0);

    const overdueCount = defects.reduce((acc, d) => {
      const dl = parseDeadline(d.deadline);
      if (!dl) return acc;
      const isFinal = d.status === 'closed' || d.status === 'canceled';
      if (!isFinal && dl < now) return acc + 1;
      return acc;
    }, 0);

    const byBuildingMap = defects.reduce<Record<number, number>>((acc, d) => {
      const bid = d.building_id || 0;
      acc[bid] = (acc[bid] || 0) + 1;
      return acc;
    }, {});

    const byBuilding = Object.entries(byBuildingMap).map(([bid, cnt]) => {
      const id = Number(bid);
      const b = buildings.find((x) => x.id === id);
      return {
        id,
        name: b ? b.name : `Здание #${id}`,
        count: cnt,
      };
    });

    return { total, countByStatus, closedCount, overdueCount, byBuilding };
  }, [defects, buildings]);

  // === EXPORT SECTION START ===

  const exportFullDefectsCsv = () => {
    const headers = [
      'ID',
      'Здание',
      'Название',
      'Описание',
      'Приоритет',
      'Статус',
      'Ответственный',
      'Дедлайн',
      'Создано',
      'Обновлено',
    ];

    const lines = [headers.join(',')];

    for (const d of defects) {
      const bName = buildings.find((b) => b.id === d.building_id)?.name || `Здание #${d.building_id}`;
      const row = [
        d.id,
        bName,
        d.title,
        d.description,
        d.priority,
        d.status,
        d.responsible_person_id,
        d.deadline,
        d.created_at,
        d.updated_at,
      ].map(csvEscape).join(',');
      lines.push(row);
    }

    const csvContent = lines.join('\r\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    saveAs(blob, `defects_full_${new Date().toISOString().slice(0,10)}.csv`);
  };

  const exportSummaryByBuildingXlsx = () => {
    const sheetData = [
      ['ID здания', 'Название', 'Всего дефектов'],
      ...metrics.byBuilding.map((b) => [b.id, b.name, b.count]),
    ];

    const ws = XLSX.utils.aoa_to_sheet(sheetData);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Сводка по зданиям');
    XLSX.writeFile(wb, `summary_by_building_${new Date().toISOString().slice(0,10)}.xlsx`);
  };

  const exportStatusBreakdownCsv = () => {
    const labels = {
      new: 'Новых',
      in_progress: 'В работе',
      review: 'На проверке',
      closed: 'Закрытых',
      canceled: 'Отменённых',
      unknown: 'Неизвестных',
    };

    const headers = ['Статус', 'Количество'];
    const lines = [headers.join(',')];

    for (const [status, count] of Object.entries(metrics.countByStatus)) {
      const label = labels[status as keyof typeof labels] || status;
      lines.push([csvEscape(label), csvEscape(count)].join(','));
    }

    const csvContent = lines.join('\r\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    saveAs(blob, `status_breakdown_${new Date().toISOString().slice(0,10)}.csv`);
  };

  // === EXPORT SECTION END ===

  const doughnutData = useMemo(() => {
    const labelMap: Record<string, string> = {
      new: 'Новых',
      in_progress: 'В работе',
      review: 'На проверке',
      closed: 'Закрытых',
      canceled: 'Отменённых',
      unknown: 'Неизвестных',
    };

    const labels = Object.keys(metrics.countByStatus).map((s) => labelMap[s] || s);
    const data = Object.keys(metrics.countByStatus).map((s) => metrics.countByStatus[s]);
    return {
      labels,
      datasets: [
        {
          data,
          backgroundColor: ['#1890ff', '#fa8c16', '#52c41a', '#ff4d4f', '#bfbfbf'],
        },
      ],
    };
  }, [metrics.countByStatus]);

  const doughnutOptions = {
    plugins: {
      legend: { position: 'bottom' as const },
      title: { display: true, text: 'Распределение дефектов по статусам' },
    },
    maintainAspectRatio: false,
  };

  const barData = useMemo(() => {
    const labels = metrics.byBuilding.map((b) => b.name);
    const data = metrics.byBuilding.map((b) => b.count);
    return {
      labels,
      datasets: [{ label: 'Дефекты', data, backgroundColor: '#1890ff' }],
    };
  }, [metrics.byBuilding]);

  const barOptions = {
    plugins: {
      legend: { display: false },
      title: { display: true, text: 'Дефекты по зданиям' },
    },
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      y: {
        beginAtZero: true,
        ticks: { stepSize: 1, precision: 0 },
      },
    },
  };

  return (
    <div className="analytics-page">
      <div className="analytics-page__back">
        <button onClick={() => navigate(-1)}>← Назад</button>
      </div>
      <h2 className="analytics-page__title">Аналитика по дефектам</h2>

      {/* === EXPORT BUTTONS === */}
      <div className="analytics-actions">
        <button onClick={exportFullDefectsCsv}>📄 Скачать все дефекты (CSV)</button>
        <button onClick={exportSummaryByBuildingXlsx}>🏢 Сводка по зданиям (XLSX)</button>
        <button onClick={exportStatusBreakdownCsv}>📊 Распределение по статусам (CSV)</button>
      </div>

      {loading ? (
        <div className="analytics-page__loading">Загрузка данных...</div>
      ) : error ? (
        <div className="analytics-page__error">{error}</div>
      ) : (
        <>
          <div className="metrics-cards">
            <div className="metric-card">
              <div className="metric-card__value">{metrics.total}</div>
              <div className="metric-card__label">Всего дефектов</div>
            </div>
            <div className="metric-card">
              <div className="metric-card__value">{metrics.countByStatus['new'] || 0}</div>
              <div className="metric-card__label">Новых</div>
            </div>
            <div className="metric-card">
              <div className="metric-card__value">{metrics.closedCount}</div>
              <div className="metric-card__label">Закрыто / Отменено</div>
            </div>
            <div className="metric-card">
              <div className="metric-card__value">{metrics.overdueCount}</div>
              <div className="metric-card__label">Просрочено</div>
            </div>
          </div>

          <div className="charts-row">
            <div className="chart-card chart-card--doughnut">
              <Doughnut data={doughnutData} options={doughnutOptions} />
            </div>

            <div className="chart-card chart-card--bar">
              <Bar data={barData} options={barOptions} />
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default AnalyticsPage;
