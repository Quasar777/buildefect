import React, { useState, useEffect } from 'react';
import { getBuildings } from '../../../api/buildings';
import './BuildingSelector.scss';

interface Building {
  id: number;
  name: string;
  address: string;
  stage: string;
}

interface BuildingSelectorProps {
  onBuildingSelect: (building: Building) => void;
  selectedBuilding: Building | null;
}

const BuildingSelector: React.FC<BuildingSelectorProps> = ({ onBuildingSelect, selectedBuilding }) => {
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [loading, setLoading] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    fetchBuildings();
  }, []);

  const fetchBuildings = async () => {
    setLoading(true);
    try {
      const data = await getBuildings(); // использует axios
      setBuildings(data);
    } catch (error) {
      console.error('Ошибка загрузки зданий:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleBuildingClick = (building: Building) => {
    onBuildingSelect(building);
    setIsOpen(false);
  };

  return (
    <div className="buildingSelector">
      <button 
        className="buildingSelector__button"
        onClick={() => setIsOpen(!isOpen)}
        disabled={loading}
      >
        {loading ? 'Загрузка...' : selectedBuilding ? selectedBuilding.name : 'Выбрать здание'}
        <span className={`buildingSelector__arrow ${isOpen ? 'buildingSelector__arrow--open' : ''}`}>
          ▼
        </span>
      </button>

      {isOpen && (
        <div className="buildingSelector__dropdown">
          {buildings.length === 0 ? (
            <div className="buildingSelector__empty">
              Нет доступных зданий
            </div>
          ) : (
            buildings.map((building) => (
              <div
                key={building.id}
                className={`buildingSelector__item ${
                  selectedBuilding?.id === building.id ? 'buildingSelector__item--selected' : ''
                }`}
                onClick={() => handleBuildingClick(building)}
              >
                <div className="buildingSelector__item-name">{building.name}</div>
                <div className="buildingSelector__item-address">{building.address}</div>
                <div className="buildingSelector__item-stage">{building.stage}</div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default BuildingSelector;
