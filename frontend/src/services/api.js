/**
 * Cliente HTTP central da API (backend-golang).
 *
 * Padrões do backend que este cliente respeita:
 * - Base URL via VITE_API_BASE_URL; vazio = mesma origem (usa o proxy do Vite em dev, evita CORS).
 * - Sucesso em JSON com campos em snake_case (ibge_code, etc.).
 * - Paginação no envelope { dados, pagina, tamanho, total }.
 * - Erros em JSON {"error":{"code","message"}}; exceção: 405 em texto puro.
 */

const RAW_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

export const API_BASE_URL = RAW_BASE_URL.replace(/\/+$/, '');

export class ApiError extends Error {
  constructor(message, { status = 0, payload = null } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

async function readBody(response) {
  const contentType = response.headers.get('content-type') || '';

  if (contentType.includes('application/json')) {
    return response.json().catch(() => null);
  }

  return response.text().then((text) => text.trim()).catch(() => '');
}

export async function request(path, { signal } = {}) {
  let response;

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'GET',
      signal,
      headers: { Accept: 'application/json' },
    });
  } catch (error) {
    if (error.name === 'AbortError') {
      throw error;
    }
    throw new ApiError('Não foi possível conectar à API. O backend está rodando?', {
      status: 0,
    });
  }

  const body = await readBody(response);

  if (!response.ok) {
    const apiMessage = typeof body?.error?.message === 'string' ? body.error.message : '';
    const message = apiMessage
      || (typeof body === 'string' && body ? body : `Erro ${response.status} na API`);
    throw new ApiError(message, { status: response.status, payload: body });
  }

  return body;
}

export const api = {
  get: (path, options) => request(path, options),
};
