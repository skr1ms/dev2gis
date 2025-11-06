import { getStoredToken, getStoredRefreshToken, setStoredToken, setStoredRefreshToken, removeStoredToken } from './auth';

interface ApiError {
  error: string;
  status?: number;
}

export class ApiException extends Error {
  status: number;
  
  constructor(message: string, status: number = 500) {
    super(message);
    this.name = 'ApiException';
    this.status = status;
  }
}

let isRefreshing = false;
let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const refreshToken = getStoredRefreshToken();
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      setStoredToken(data.access_token);
      setStoredRefreshToken(data.refresh_token);
      
      return data.access_token;
    } catch (error) {
      removeStoredToken();
      if (typeof window !== 'undefined') {
        window.location.href = '/auth/login';
      }
      throw error;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

function extractErrorMessage(data: any, status: number): string {
  if (data.error) {
    return data.error;
  }
  if (data.message) {
    return data.message;
  }
  if (data.detail) {
    return data.detail;
  }
  
  switch (status) {
    case 400:
      return 'Некорректные данные. Проверьте введенную информацию.';
    case 401:
      return 'Неверный email или пароль';
    case 403:
      return 'Доступ запрещен';
    case 404:
      return 'Ресурс не найден';
    case 409:
      return 'Пользователь с таким email уже существует';
    case 422:
      return 'Ошибка валидации данных';
    case 429:
      return 'Слишком много запросов. Попробуйте позже.';
    case 500:
      return 'Ошибка сервера. Попробуйте позже.';
    case 503:
      return 'Сервис временно недоступен';
    default:
      return `Ошибка: ${status}`;
  }
}

export async function apiRequest<T = any>(
  url: string,
  options: RequestInit = {},
  retryCount = 0
): Promise<T> {
  const token = getStoredToken();
  
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> || {}),
  };
  
  if (token && !headers['Authorization']) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  if (!(options.body instanceof FormData) && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json';
  }
  
  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });
    
    let data: any = {};
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        data = await response.json();
      } catch (e) {
        console.warn('Failed to parse JSON response:', e);
      }
    }
    
    if (!response.ok) {
      if (response.status === 401 && retryCount === 0 && !url.includes('/auth/refresh') && !url.includes('/auth/login') && !url.includes('/auth/register')) {
        try {
          await refreshAccessToken();
          return apiRequest<T>(url, options, retryCount + 1);
        } catch (refreshError) {
          throw new ApiException('Сессия истекла, пожалуйста войдите снова', 401);
        }
      }
      
      const errorMessage = extractErrorMessage(data, response.status);
      throw new ApiException(errorMessage, response.status);
    }
    
    return data as T;
  } catch (error) {
    if (error instanceof ApiException) {
      throw error;
    }
    
    if (error instanceof TypeError && error.message === 'Failed to fetch') {
      throw new ApiException('Ошибка сети. Проверьте подключение к интернету.', 0);
    }
    
    throw new ApiException(
      error instanceof Error ? error.message : 'Неизвестная ошибка',
      0
    );
  }
}

export const api = {
  get: <T = any>(url: string, options?: RequestInit) =>
    apiRequest<T>(url, { ...options, method: 'GET' }),
  
  post: <T = any>(url: string, body?: any, options?: RequestInit) =>
    apiRequest<T>(url, {
      ...options,
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),
  
  put: <T = any>(url: string, body?: any, options?: RequestInit) =>
    apiRequest<T>(url, {
      ...options,
      method: 'PUT',
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),
  
  delete: <T = any>(url: string, options?: RequestInit) =>
    apiRequest<T>(url, { ...options, method: 'DELETE' }),
};

