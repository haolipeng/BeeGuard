import axios, { AxiosError } from 'axios';

const request = axios.create({
  baseURL: '/api1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // 登录/登出接口的 401 不做跳转，交给页面自己处理错误提示
      const url = error.config?.url || '';
      if (!url.includes('/auth/')) {
        localStorage.removeItem('token');
        window.location.href = '/ui/login';
      }
    }
    return Promise.reject(error);
  }
);

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export default request;
