import { User, LoginRequest, RegisterRequest } from '@rwa-platform/types';
import apiClient from './client';

export interface AuthResponse {
  user: User;
  tokens: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
}

export interface Web3LoginRequest {
  address: string;
  signature: string;
  message: string;
  chainId: number;
}

export interface EmailVerificationRequest {
  email: string;
  code: string;
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirmRequest {
  token: string;
  newPassword: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

class AuthAPI {
  /**
   * 邮箱密码登录
   */
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    return apiClient.post('/auth/login', credentials);
  }

  /**
   * Web3钱包登录
   */
  async web3Login(request: Web3LoginRequest): Promise<AuthResponse> {
    return apiClient.post('/auth/web3-login', request);
  }

  /**
   * 注册新用户
   */
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return apiClient.post('/auth/register', data);
  }

  /**
   * 刷新访问令牌
   */
  async refreshToken(refreshToken: string): Promise<AuthResponse> {
    return apiClient.post('/auth/refresh', { refreshToken });
  }

  /**
   * 登出
   */
  async logout(): Promise<void> {
    return apiClient.post('/auth/logout');
  }

  /**
   * 获取当前用户信息
   */
  async getCurrentUser(): Promise<User> {
    return apiClient.get('/auth/me');
  }

  /**
   * 发送邮箱验证码
   */
  async sendEmailVerification(email: string): Promise<void> {
    return apiClient.post('/auth/send-verification', { email });
  }

  /**
   * 验证邮箱
   */
  async verifyEmail(request: EmailVerificationRequest): Promise<void> {
    return apiClient.post('/auth/verify-email', request);
  }

  /**
   * 发送密码重置邮件
   */
  async requestPasswordReset(request: PasswordResetRequest): Promise<void> {
    return apiClient.post('/auth/forgot-password', request);
  }

  /**
   * 重置密码
   */
  async resetPassword(request: PasswordResetConfirmRequest): Promise<void> {
    return apiClient.post('/auth/reset-password', request);
  }

  /**
   * 修改密码
   */
  async changePassword(request: ChangePasswordRequest): Promise<void> {
    return apiClient.post('/auth/change-password', request);
  }

  /**
   * 获取Web3登录消息
   */
  async getWeb3LoginMessage(address: string, chainId: number): Promise<{ message: string; nonce: string }> {
    return apiClient.post('/auth/web3-message', { address, chainId });
  }

  /**
   * 绑定Web3钱包
   */
  async linkWallet(address: string, signature: string, message: string, chainId: number): Promise<void> {
    return apiClient.post('/auth/link-wallet', {
      address,
      signature,
      message,
      chainId,
    });
  }

  /**
   * 解绑Web3钱包
   */
  async unlinkWallet(address: string): Promise<void> {
    return apiClient.delete(`/auth/unlink-wallet/${address}`);
  }

  /**
   * 获取用户绑定的钱包列表
   */
  async getLinkedWallets(): Promise<Array<{
    address: string;
    chain: string;
    linkedAt: string;
    isActive: boolean;
  }>> {
    return apiClient.get('/auth/linked-wallets');
  }

  /**
   * 启用两步验证
   */
  async enableTwoFactor(): Promise<{
    secret: string;
    qrCode: string;
    backupCodes: string[];
  }> {
    return apiClient.post('/auth/2fa/enable');
  }

  /**
   * 确认启用两步验证
   */
  async confirmTwoFactor(code: string): Promise<void> {
    return apiClient.post('/auth/2fa/confirm', { code });
  }

  /**
   * 禁用两步验证
   */
  async disableTwoFactor(code: string): Promise<void> {
    return apiClient.post('/auth/2fa/disable', { code });
  }

  /**
   * 验证两步验证码
   */
  async verifyTwoFactor(code: string): Promise<void> {
    return apiClient.post('/auth/2fa/verify', { code });
  }

  /**
   * 生成新的备份码
   */
  async generateBackupCodes(): Promise<string[]> {
    return apiClient.post('/auth/2fa/backup-codes');
  }

  /**
   * 检查用户名是否可用
   */
  async checkUsernameAvailability(username: string): Promise<{ available: boolean }> {
    return apiClient.get(`/auth/check-username/${username}`);
  }

  /**
   * 检查邮箱是否已注册
   */
  async checkEmailExists(email: string): Promise<{ exists: boolean }> {
    return apiClient.get(`/auth/check-email/${email}`);
  }

  /**
   * 获取用户会话列表
   */
  async getUserSessions(): Promise<Array<{
    id: string;
    deviceInfo: string;
    ipAddress: string;
    location: string;
    lastActiveAt: string;
    isCurrent: boolean;
  }>> {
    return apiClient.get('/auth/sessions');
  }

  /**
   * 终止指定会话
   */
  async terminateSession(sessionId: string): Promise<void> {
    return apiClient.delete(`/auth/sessions/${sessionId}`);
  }

  /**
   * 终止所有其他会话
   */
  async terminateAllOtherSessions(): Promise<void> {
    return apiClient.delete('/auth/sessions/others');
  }

  /**
   * 获取登录历史
   */
  async getLoginHistory(page: number = 1, limit: number = 20): Promise<{
    data: Array<{
      id: string;
      ipAddress: string;
      userAgent: string;
      location: string;
      success: boolean;
      loginMethod: string;
      timestamp: string;
    }>;
    total: number;
    page: number;
    totalPages: number;
  }> {
    return apiClient.get('/auth/login-history', {
      params: { page, limit },
    });
  }

  /**
   * 更新用户资料
   */
  async updateProfile(data: {
    firstName?: string;
    lastName?: string;
    displayName?: string;
    avatar?: string;
    dateOfBirth?: string;
    nationality?: string;
    phone?: string;
  }): Promise<User> {
    return apiClient.patch('/auth/profile', data);
  }

  /**
   * 上传头像
   */
  async uploadAvatar(file: File): Promise<{ url: string }> {
    return apiClient.upload('/auth/avatar', file);
  }

  /**
   * 删除账户
   */
  async deleteAccount(password: string, reason?: string): Promise<void> {
    return apiClient.delete('/auth/account', {
      data: { password, reason },
    });
  }

  /**
   * 导出用户数据
   */
  async exportUserData(): Promise<void> {
    return apiClient.download('/auth/export-data', 'user-data.json');
  }
}

export const authApi = new AuthAPI();
