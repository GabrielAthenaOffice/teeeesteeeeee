import React from 'react';
import { useLocation } from 'react-router-dom';
import { ApiStatus } from '../components/ApiStatus';
import './TopBar.css';

const MODULE_TITLES = {
  '/': 'Dashboard',
  '/estados': 'Estados',
  '/cidades': 'Cidades',
};

export function TopBar() {
  const location = useLocation();
  const moduleTitle =
    MODULE_TITLES[location.pathname]
    ?? (location.pathname.startsWith('/cidades/') ? 'Detalhe da cidade' : 'TRADS');

  return (
    <header className="topbar">
      <div className="topbar-left">
        <div className="topbar-module-title">{moduleTitle}</div>
      </div>

      <div className="topbar-right">
        <span className="topbar-env">{import.meta.env.MODE}</span>
        <ApiStatus />
      </div>
    </header>
  );
}
