import { UUID, Timestamp, Region, Currency, AddressInfo } from './common';

// 用户类型
export enum UserType {
  INDIVIDUAL = 'individual',
  INSTITUTIONAL = 'institutional',
  FAMILY_OFFICE = 'family_office'
}

// KYC状态
export enum KYCStatus {
  NOT_STARTED = 'not_started',
  PENDING = 'pending',
  APPROVED = 'approved',
  REJECTED = 'rejected',
  EXPIRED = 'expired'
}

// KYC等级
export enum KYCLevel {
  BASIC = 'basic',      // 基础认证
  STANDARD = 'standard', // 标准认证
  ENHANCED = 'enhanced'  // 增强认证
}

// 用户基础信息
export interface User {
  id: UUID;
  email: string;
  phone?: string;
  
  // 个人信息
  profile: {
    firstName?: string;
    lastName?: string;
    displayName?: string;
    avatar?: string;
    dateOfBirth?: string;
    nationality?: Region;
    residenceCountry: Region;
  };
  
  // 用户类型和状态
  type: UserType;
  status: 'active' | 'suspended' | 'deactivated';
  
  // KYC信息
  kyc: {
    status: KYCStatus;
    level: KYCLevel;
    completedAt?: Timestamp;
    expiresAt?: Timestamp;
    documents: {
      type: string;
      status: 'pending' | 'approved' | 'rejected';
      uploadedAt: Timestamp;
    }[];
  };
  
  // 合规信息
  compliance: {
    isAccredited: boolean;
    sanctionsCheck: {
      status: 'clear' | 'flagged' | 'pending';
      checkedAt: Timestamp;
    };
    pepCheck: {
      status: 'clear' | 'flagged' | 'pending';
      checkedAt: Timestamp;
    };
    riskLevel: 'low' | 'medium' | 'high';
  };
  
  // 偏好设置
  preferences: {
    language: string;
    currency: Currency;
    timezone: string;
    notifications: {
      email: boolean;
      sms: boolean;
      push: boolean;
      telegram?: boolean;
    };
    privacy: {
      shareData: boolean;
      marketing: boolean;
    };
  };
  
  // 时间戳
  createdAt: Timestamp;
  updatedAt: Timestamp;
  lastLoginAt?: Timestamp;
}

// 用户钱包/账户绑定
export interface UserAccount {
  id: UUID;
  userId: UUID;
  
  // 账户类型
  type: 'wallet' | 'exchange' | 'broker' | 'bank';
  
  // 账户信息
  account: {
    // 钱包地址
    address?: AddressInfo;
    
    // 交易所/券商
    platform?: string;
    accountId?: string;
    
    // 银行账户
    bankName?: string;
    accountNumber?: string;
    routingNumber?: string;
  };
  
  // 权限和状态
  permissions: {
    read: boolean;
    trade?: boolean;
  };
  
  status: 'active' | 'inactive' | 'error';
  
  // 连接信息
  connection: {
    method: 'oauth' | 'api_key' | 'read_only';
    connectedAt: Timestamp;
    lastSyncAt?: Timestamp;
    syncStatus: 'success' | 'error' | 'pending';
    errorMessage?: string;
  };
  
  // 元数据
  metadata: {
    label?: string;
    notes?: string;
  };
  
  createdAt: Timestamp;
  updatedAt: Timestamp;
}

// 用户会话
export interface UserSession {
  id: UUID;
  userId: UUID;
  
  // 会话信息
  sessionToken: string;
  refreshToken?: string;
  
  // 设备信息
  device: {
    userAgent: string;
    ip: string;
    location?: {
      country: string;
      city?: string;
    };
  };
  
  // 时间信息
  createdAt: Timestamp;
  expiresAt: Timestamp;
  lastActiveAt: Timestamp;
}

// 用户活动日志
export interface UserActivity {
  id: UUID;
  userId: UUID;
  
  // 活动类型
  type: 'login' | 'logout' | 'view_asset' | 'connect_account' | 'trade_redirect' | 'kyc_update';
  
  // 活动详情
  details: {
    action: string;
    resource?: string;
    metadata?: Record<string, any>;
  };
  
  // 上下文信息
  context: {
    ip: string;
    userAgent: string;
    sessionId?: UUID;
  };
  
  timestamp: Timestamp;
}

// 用户通知
export interface UserNotification {
  id: UUID;
  userId: UUID;
  
  // 通知类型
  type: 'price_alert' | 'risk_warning' | 'kyc_update' | 'system_announcement' | 'portfolio_update';
  
  // 通知内容
  title: string;
  message: string;
  data?: Record<string, any>;
  
  // 通知状态
  status: 'unread' | 'read' | 'archived';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  
  // 发送渠道
  channels: {
    email?: boolean;
    sms?: boolean;
    push?: boolean;
    inApp: boolean;
  };
  
  // 时间信息
  createdAt: Timestamp;
  readAt?: Timestamp;
  expiresAt?: Timestamp;
}
