/**
 * Token存储工具类
 * 处理JWT token的安全存储和管理
 */

const ACCESS_TOKEN_KEY = 'rwa_access_token';
const REFRESH_TOKEN_KEY = 'rwa_refresh_token';
const TOKEN_EXPIRY_KEY = 'rwa_token_expiry';

class TokenStorage {
  /**
   * 检查是否在浏览器环境
   */
  private isBrowser(): boolean {
    return typeof window !== 'undefined';
  }

  /**
   * 设置访问令牌
   */
  setAccessToken(token: string, expiresIn?: number): void {
    if (!this.isBrowser()) return;

    try {
      localStorage.setItem(ACCESS_TOKEN_KEY, token);
      
      if (expiresIn) {
        const expiryTime = Date.now() + expiresIn * 1000;
        localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
      }
    } catch (error) {
      console.error('Failed to set access token:', error);
    }
  }

  /**
   * 获取访问令牌
   */
  getAccessToken(): string | null {
    if (!this.isBrowser()) return null;

    try {
      const token = localStorage.getItem(ACCESS_TOKEN_KEY);
      
      // 检查token是否过期
      if (token && this.isTokenExpired()) {
        this.clearTokens();
        return null;
      }
      
      return token;
    } catch (error) {
      console.error('Failed to get access token:', error);
      return null;
    }
  }

  /**
   * 设置刷新令牌
   */
  setRefreshToken(token: string): void {
    if (!this.isBrowser()) return;

    try {
      localStorage.setItem(REFRESH_TOKEN_KEY, token);
    } catch (error) {
      console.error('Failed to set refresh token:', error);
    }
  }

  /**
   * 获取刷新令牌
   */
  getRefreshToken(): string | null {
    if (!this.isBrowser()) return null;

    try {
      return localStorage.getItem(REFRESH_TOKEN_KEY);
    } catch (error) {
      console.error('Failed to get refresh token:', error);
      return null;
    }
  }

  /**
   * 同时设置访问令牌和刷新令牌
   */
  setTokens(accessToken: string, refreshToken: string, expiresIn?: number): void {
    this.setAccessToken(accessToken, expiresIn);
    this.setRefreshToken(refreshToken);
  }

  /**
   * 清除所有令牌
   */
  clearTokens(): void {
    if (!this.isBrowser()) return;

    try {
      localStorage.removeItem(ACCESS_TOKEN_KEY);
      localStorage.removeItem(REFRESH_TOKEN_KEY);
      localStorage.removeItem(TOKEN_EXPIRY_KEY);
    } catch (error) {
      console.error('Failed to clear tokens:', error);
    }
  }

  /**
   * 检查访问令牌是否存在
   */
  hasAccessToken(): boolean {
    return !!this.getAccessToken();
  }

  /**
   * 检查刷新令牌是否存在
   */
  hasRefreshToken(): boolean {
    return !!this.getRefreshToken();
  }

  /**
   * 检查token是否过期
   */
  isTokenExpired(): boolean {
    if (!this.isBrowser()) return true;

    try {
      const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);
      if (!expiryTime) return false; // 如果没有过期时间，假设未过期
      
      return Date.now() >= parseInt(expiryTime);
    } catch (error) {
      console.error('Failed to check token expiry:', error);
      return true;
    }
  }

  /**
   * 获取token过期时间
   */
  getTokenExpiry(): Date | null {
    if (!this.isBrowser()) return null;

    try {
      const expiryTime = localStorage.getItem(TOKEN_EXPIRY_KEY);
      return expiryTime ? new Date(parseInt(expiryTime)) : null;
    } catch (error) {
      console.error('Failed to get token expiry:', error);
      return null;
    }
  }

  /**
   * 获取token剩余有效时间（秒）
   */
  getTokenRemainingTime(): number {
    const expiry = this.getTokenExpiry();
    if (!expiry) return 0;
    
    const remaining = Math.max(0, expiry.getTime() - Date.now());
    return Math.floor(remaining / 1000);
  }

  /**
   * 检查token是否即将过期（默认5分钟内）
   */
  isTokenExpiringSoon(thresholdMinutes: number = 5): boolean {
    const remainingSeconds = this.getTokenRemainingTime();
    return remainingSeconds <= thresholdMinutes * 60;
  }

  /**
   * 从JWT token中解析payload（不验证签名）
   */
  parseJwtPayload(token: string): any {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      return JSON.parse(jsonPayload);
    } catch (error) {
      console.error('Failed to parse JWT payload:', error);
      return null;
    }
  }

  /**
   * 获取当前用户ID（从token中解析）
   */
  getCurrentUserId(): string | null {
    const token = this.getAccessToken();
    if (!token) return null;

    const payload = this.parseJwtPayload(token);
    return payload?.sub || payload?.userId || null;
  }

  /**
   * 获取当前用户角色（从token中解析）
   */
  getCurrentUserRoles(): string[] {
    const token = this.getAccessToken();
    if (!token) return [];

    const payload = this.parseJwtPayload(token);
    return payload?.roles || payload?.authorities || [];
  }

  /**
   * 检查用户是否有特定角色
   */
  hasRole(role: string): boolean {
    const roles = this.getCurrentUserRoles();
    return roles.includes(role);
  }

  /**
   * 检查用户是否有任一角色
   */
  hasAnyRole(roles: string[]): boolean {
    const userRoles = this.getCurrentUserRoles();
    return roles.some(role => userRoles.includes(role));
  }

  /**
   * 检查用户是否有所有角色
   */
  hasAllRoles(roles: string[]): boolean {
    const userRoles = this.getCurrentUserRoles();
    return roles.every(role => userRoles.includes(role));
  }

  /**
   * 获取token信息摘要
   */
  getTokenInfo(): {
    hasAccessToken: boolean;
    hasRefreshToken: boolean;
    isExpired: boolean;
    isExpiringSoon: boolean;
    remainingTime: number;
    userId: string | null;
    roles: string[];
  } {
    return {
      hasAccessToken: this.hasAccessToken(),
      hasRefreshToken: this.hasRefreshToken(),
      isExpired: this.isTokenExpired(),
      isExpiringSoon: this.isTokenExpiringSoon(),
      remainingTime: this.getTokenRemainingTime(),
      userId: this.getCurrentUserId(),
      roles: this.getCurrentUserRoles(),
    };
  }
}

// 导出单例实例
export const tokenStorage = new TokenStorage();
