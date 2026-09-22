import { api } from './api';

export const DEFAULT_PAGE_SIZE = 20;
export const PAGE_SIZE_OPTIONS = [20, 50, 100];

/**
 * GET /api/v1/states/{ibgeCode}/cities?page=1&pageSize=20
 *
 * Resposta no envelope de paginação do backend:
 * { dados: [{ id, ibge_code, name, indicators }], pagina, tamanho, total }
 * indicators: { population, income, gdp }, cada item { year, value } ou null.
 *
 * @param {number|string} stateIbgeCode código IBGE da UF (path param)
 * @returns {Promise<{ dados: Array, pagina: number, tamanho: number, total: number }>}
 */
export async function getCities(stateIbgeCode, { page = 1, pageSize = DEFAULT_PAGE_SIZE, signal } = {}) {
  if (!stateIbgeCode) {
    return { dados: [], pagina: 1, tamanho: pageSize, total: 0 };
  }

  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });

  const response = await api.get(
    `/api/v1/states/${stateIbgeCode}/cities?${params.toString()}`,
    { signal },
  );

  return {
    dados: response?.dados ?? [],
    pagina: response?.pagina ?? page,
    tamanho: response?.tamanho ?? pageSize,
    total: response?.total ?? 0,
  };
}
