import { Routes, Route, Navigate } from 'react-router-dom';
import { AppShell } from './app/layout/AppShell';

import Dashboard from './modules/dashboard/Dashboard';
import States from './modules/states/States';
import Cities from './modules/cities/Cities';
import CityDetail from './modules/cities/CityDetail';

function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/estados" element={<States />} />
        <Route path="/cidades" element={<Cities />} />
        <Route path="/cidades/:ibgeCode" element={<CityDetail />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  );
}

export default App;
