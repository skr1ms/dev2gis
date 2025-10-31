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
    
    const data = await response.json().catch(() => ({}));
    
    if (!response.ok) {
      if (response.status === 401 && retryCount === 0 && !url.includes('/auth/refresh')) {
        try {
          await refreshAccessToken();
          return apiRequest<T>(url, options, retryCount + 1);
        } catch (refreshError) {
          throw new ApiException('Сессия истекла, пожалуйста войдите снова', 401);
        }
      }
      
      throw new ApiException(
        data.error || `HTTP error! status: ${response.status}`,
        response.status
      );
    }
    
    return data as T;
  } catch (error) {
    if (error instanceof ApiException) {
      throw error;
    }
    throw new ApiException(
      error instanceof Error ? error.message : 'Network error',
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

