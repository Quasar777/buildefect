// src/App.tsx
import React, { Suspense, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import SignInPage from './pages/SignInPage/SignInPage';
import SignUpPage from './pages/SignUpPage/SignUpPage';
import BuildingSelector from './components/UI/BuildingSelector/BuildingSelector';
import DefectsList from './components/UI/DefectsList/DefectsList';
import { RequireAuth, RequireRole } from './components/RouteGuards';
import './App.scss';

// Ленивая загрузка страниц (создай файлы, если их ещё нет)
const DefectPage = React.lazy(() => import('./pages/DefectPage/DefectPage').catch(() => ({ default: () => <div>Ошибка загрузки страницы дефекта</div> })));
const NewDefectPage = React.lazy(() => import('./pages/CreateDefectPage/CreateDefectPage').catch(() => ({ default: () => <div>Ошибка загрузки страницы создания дефекта</div> })));
const AnalyticsPage = React.lazy(() => import('./pages/AnalyticsPage/AnalyticsPage').catch(() => ({ default: () => <div>Ошибка загрузки аналитики</div> })));

type Building = {
  id: number;
  name: string;
  address: string;
  stage: string;
};

const PrivateLayout: React.FC<{ children: React.ReactNode; selectedBuilding: Building | null; setSelectedBuilding: (b: Building | null) => void; }> = ({ children, selectedBuilding, setSelectedBuilding }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="App">
      <header className="header">
        <h1 className="app-title">BuilDefect</h1>
        <div className="header__right">
          <BuildingSelector
            onBuildingSelect={setSelectedBuilding}
            selectedBuilding={selectedBuilding}
          />
          <button
            className="header__btn"
            onClick={() => navigate('/defects/new')}
            title="Создать дефект"
          >
            ➕ Новый дефект
          </button>
          <button
            className="header__btn"
            onClick={() => navigate('/analytics')}
            title="Аналитика"
            disabled={user?.role !== 'observer'}
          >
            📊 Аналитика
          </button>

          <div className="user-info">
            <span className="user-name">Привет, {user?.name} {user?.lastname}</span>
            <span className="user-role">{user?.role} | ID: {user?.id}</span>
            <button className="logout-btn" onClick={() => logout()}>
              Выйти
            </button>
          </div>
        </div>
      </header>

      <main className="main">
        {children}
      </main>
    </div>
  );
};

const Dashboard: React.FC = () => {
  const [selectedBuilding, setSelectedBuilding] = useState<Building | null>(null);

  return (
    <PrivateLayout selectedBuilding={selectedBuilding} setSelectedBuilding={setSelectedBuilding}>
      <div className="dashboard">
        <DefectsList
          buildingId={selectedBuilding?.id || null}
          onDefectClick={(defect) => {
            // при клике — переход на страницу дефекта
            // предполагается, что defect имеет поле id
            window.location.href = `/defects/${defect.id}`;
          }}
        />
      </div>
    </PrivateLayout>
  );
};

const AppRoutes: React.FC = () => {
  const { isAuth } = useAuth();

  return (
    <Suspense fallback={<div>Загрузка...</div>}>
      <Routes>
        <Route path="/signin" element={!isAuth ? <SignInPage /> : <Navigate to="/" replace />} />
        <Route path="/signup" element={!isAuth ? <SignUpPage /> : <Navigate to="/" replace />} />

        {/* Приватные маршруты */}
        <Route
          path="/"
          element={
            <RequireAuth>
              <Dashboard />
            </RequireAuth>
          }
        />

        <Route
          path="/defects/new"
          element={
            <RequireAuth>
              {/* любой авторизованный может создать дефект (engineer/manager/observer могут по API, но можно ограничить) */}
              <NewDefectPage />
            </RequireAuth>
          }
        />

        <Route
          path="/defects/:id"
          element={
            <RequireAuth>
              <DefectPage />
            </RequireAuth>
          }
        />

        <Route
          path="/analytics"
          element={
            <RequireAuth>
              <RequireRole roles={['observer']}>
                <AnalyticsPage />
              </RequireRole>
            </RequireAuth>
          }
        />

        {/* catch-all */}
        <Route path="*" element={<Navigate to={isAuth ? '/' : '/signin'} replace />} />
      </Routes>
    </Suspense>
  );
};

const AppContent: React.FC = () => {
  const { loading } = useAuth();

  if (loading) {
    return (
      <div className="App">
        <div className="loading">
          <p>Загрузка...</p>
        </div>
      </div>
    );
  }

  return <AppRoutes />;
};

function App() {
  return (
    <AuthProvider>
      <Router>
        <AppContent />
      </Router>
    </AuthProvider>
  );
}

export default App;
