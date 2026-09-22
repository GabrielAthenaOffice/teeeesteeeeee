import { api } from './api';

export function getNationalMetrics({ signal } = {}) {
  return api.get('/api/v1/dashboard/national', { signal });
}

export function getStateMetrics({ signal } = {}) {
  return api.get('/api/v1/dashboard/states', { signal });
}

export function getTopCities({ signal } = {}) {
  return api.get('/api/v1/dashboard/top-cities', { signal });
}

export function getAgeDistribution({ signal } = {}) {
  return api.get('/api/v1/dashboard/age', { signal });
}
