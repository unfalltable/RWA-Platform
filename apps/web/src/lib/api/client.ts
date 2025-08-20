import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { toast } from 'react-hot-toast';
import { tokenStorage } from '@/lib/utils/tokenStorage';

// API响应类型
export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
  timestamp: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: any;
  timestamp: string;
  path?: string;
}

// 请求配置
interface ApiClientConfig {
  baseURL: string;
  timeout: number;
  withCredentials: boolean;
}

class ApiClient {
  private client: AxiosInstance;
  private isRefreshing = false;
  private failedQueue: Array<{
    resolve: (value: any) => void;
    reject: (error: any) => void;
  }> = [];

  constructor(config: ApiClientConfig) {
    this.client = axios.create(config);
    this.setupInterceptors();
  }

  private setupInterceptors() {
    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        // 添加认证头
        const token = tokenStorage.getAccessToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 添加请求ID用于追踪
        config.headers['X-Request-ID'] = this.generateRequestId();

        // 添加语言头
        const language = this.getLanguage();
        if (language) {
          config.headers['Accept-Language'] = language;
        }

        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.client.interceptors.response.use(
      (response) => {
        return response;
      },
      async (error: AxiosError) => {
        const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };

        // 处理401错误（token过期）
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            // 如果正在刷新token，将请求加入队列
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject });
            }).then((token) => {
              originalRequest.headers!.Authorization = `Bearer ${token}`;
              return this.client(originalRequest);
            }).catch((err) => {
              return Promise.reject(err);
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const refreshToken = tokenStorage.getRefreshToken();
            if (!refreshToken) {
              throw new Error('No refresh token available');
            }

            // 刷新token
            const response = await this.refreshToken(refreshToken);
            const { accessToken, refreshToken: newRefreshToken } = response.tokens;

            // 更新存储的token
            tokenStorage.setTokens(accessToken, newRefreshToken);

            // 处理队列中的请求
            this.processQueue(null, accessToken);

            // 重试原始请求
            originalRequest.headers!.Authorization = `Bearer ${accessToken}`;
            return this.client(originalRequest);

          } catch (refreshError) {
            // 刷新失败，清除token并重定向到登录页
            this.processQueue(refreshError, null);
            tokenStorage.clearTokens();
            
            // 只在浏览器环境中重定向
            if (typeof window !== 'undefined') {
              window.location.href = '/login';
            }
            
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        // 处理其他错误
        this.handleError(error);
        return Promise.reject(error);
      }
    );
  }

  private processQueue(error: any, token: string | null) {
    this.failedQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve(token);
      }
    });
    
    this.failedQueue = [];
  }

  private async refreshToken(refreshToken: string) {
    const response = await axios.post('/api/auth/refresh', {
      refreshToken,
    });
    return response.data;
  }

  private handleError(error: AxiosError) {
    const apiError = error.response?.data as ApiError;
    
    // 根据错误类型显示不同的提示
    if (error.response?.status === 400) {
      toast.error(apiError?.message || '请求参数错误');
    } else if (error.response?.status === 403) {
      toast.error('权限不足，无法访问该资源');
    } else if (error.response?.status === 404) {
      toast.error('请求的资源不存在');
    } else if (error.response?.status === 429) {
      toast.error('请求过于频繁，请稍后再试');
    } else if (error.response?.status >= 500) {
      toast.error('服务器错误，请稍后再试');
    } else if (error.code === 'NETWORK_ERROR') {
      toast.error('网络连接失败，请检查网络设置');
    } else if (error.code === 'ECONNABORTED') {
      toast.error('请求超时，请稍后再试');
    }
  }

  private generateRequestId(): string {
    return Math.random().toString(36).substr(2, 9);
  }

  private getLanguage(): string {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('language') || navigator.language || 'zh-CN';
    }
    return 'zh-CN';
  }

  // HTTP方法封装
  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<ApiResponse<T>>(url, config);
    return response.data.data;
  }

  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  async patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.patch<ApiResponse<T>>(url, data, config);
    return response.data.data;
  }

  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<ApiResponse<T>>(url, config);
    return response.data.data;
  }

  // 文件上传
  async upload<T = any>(
    url: string, 
    file: File, 
    onProgress?: (progress: number) => void
  ): Promise<T> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await this.client.post<ApiResponse<T>>(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });

    return response.data.data;
  }

  // 下载文件
  async download(url: string, filename?: string): Promise<void> {
    const response = await this.client.get(url, {
      responseType: 'blob',
    });

    // 创建下载链接
    const blob = new Blob([response.data]);
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename || 'download';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  }

  // 获取原始axios实例（用于特殊需求）
  getAxiosInstance(): AxiosInstance {
    return this.client;
  }
}

// 创建API客户端实例
const apiClient = new ApiClient({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api',
  timeout: 30000,
  withCredentials: true,
});

export default apiClient;
