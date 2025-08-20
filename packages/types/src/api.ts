import { UUID, Timestamp, PaginationParams, PaginatedResponse, ApiResponse } from './common';
import { Asset, AssetMetrics, AssetRating, AssetFilter } from './asset';
import { User, UserAccount, UserNotification } from './user';
import { Channel, ChannelRating, ChannelMatch, ChannelFilter } from './channel';
import { Position, Portfolio, Cashflow, PositionFilter } from './position';
import { RiskAssessment, RiskAlert, ComplianceCheck } from './risk';

// API端点定义
export interface ApiEndpoints {
  // 认证相关
  auth: {
    login: '/auth/login';
    logout: '/auth/logout';
    refresh: '/auth/refresh';
    register: '/auth/register';
    verify: '/auth/verify';
    resetPassword: '/auth/reset-password';
  };
  
  // 用户相关
  users: {
    profile: '/users/profile';
    accounts: '/users/accounts';
    notifications: '/users/notifications';
    preferences: '/users/preferences';
    kyc: '/users/kyc';
  };
  
  // 资产相关
  assets: {
    list: '/assets';
    detail: '/assets/:id';
    metrics: '/assets/:id/metrics';
    rating: '/assets/:id/rating';
    compare: '/assets/compare';
    search: '/assets/search';
  };
  
  // 渠道相关
  channels: {
    list: '/channels';
    detail: '/channels/:id';
    rating: '/channels/:id/rating';
    match: '/channels/match';
  };
  
  // 持仓相关
  portfolio: {
    overview: '/portfolio';
    positions: '/portfolio/positions';
    history: '/portfolio/history';
    cashflow: '/portfolio/cashflow';
    reports: '/portfolio/reports';
    sync: '/portfolio/sync';
  };
  
  // 风控相关
  risk: {
    assessment: '/risk/assessment';
    alerts: '/risk/alerts';
    compliance: '/risk/compliance';
    events: '/risk/events';
  };
}

// 请求类型定义
export interface LoginRequest {
  email: string;
  password: string;
  rememberMe?: boolean;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName?: string;
  lastName?: string;
  referralCode?: string;
}

export interface AssetListRequest extends PaginationParams {
  filter?: AssetFilter;
}

export interface AssetCompareRequest {
  assetIds: UUID[];
  metrics?: string[];
}

export interface ChannelMatchRequest {
  assetId: UUID;
  amount?: number;
  region?: string;
  kycLevel?: string;
  paymentMethod?: string;
}

export interface PositionListRequest extends PaginationParams {
  filter?: PositionFilter;
}

export interface PortfolioSyncRequest {
  accountIds?: UUID[];
  force?: boolean;
}

export interface RiskAlertRequest extends PaginationParams {
  status?: string[];
  level?: string[];
  type?: string[];
}

// 响应类型定义
export interface LoginResponse {
  user: User;
  tokens: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
}

export interface AssetListResponse extends PaginatedResponse<Asset> {
  filters: {
    types: string[];
    chains: string[];
    regions: string[];
  };
}

export interface AssetDetailResponse {
  asset: Asset;
  metrics: AssetMetrics;
  rating: AssetRating;
  relatedAssets: Asset[];
}

export interface AssetCompareResponse {
  assets: Asset[];
  comparison: {
    metrics: {
      [assetId: string]: AssetMetrics;
    };
    ratings: {
      [assetId: string]: AssetRating;
    };
  };
}

export interface ChannelMatchResponse {
  matches: ChannelMatch[];
  recommendations: {
    best: ChannelMatch;
    cheapest: ChannelMatch;
    fastest: ChannelMatch;
  };
}

export interface PortfolioResponse {
  portfolio: Portfolio;
  positions: Position[];
  recentActivity: Cashflow[];
}

export interface PositionListResponse extends PaginatedResponse<Position> {
  summary: {
    totalValue: number;
    totalPnl: number;
    totalYield: number;
  };
}

export interface RiskAssessmentResponse {
  assessment: RiskAssessment;
  alerts: RiskAlert[];
  recommendations: string[];
}

// GraphQL查询类型
export interface GraphQLQueries {
  // 资产查询
  assets: {
    query: string;
    variables: {
      filter?: AssetFilter;
      pagination?: PaginationParams;
    };
  };
  
  // 用户查询
  user: {
    query: string;
    variables: {
      id?: UUID;
    };
  };
  
  // 持仓查询
  portfolio: {
    query: string;
    variables: {
      userId: UUID;
      filter?: PositionFilter;
    };
  };
}

// WebSocket事件类型
export interface WebSocketEvents {
  // 价格更新
  'price:update': {
    assetId: UUID;
    price: number;
    change24h: number;
    timestamp: Timestamp;
  };
  
  // 持仓更新
  'position:update': {
    positionId: UUID;
    changes: Partial<Position>;
    timestamp: Timestamp;
  };
  
  // 风险预警
  'risk:alert': {
    alert: RiskAlert;
    timestamp: Timestamp;
  };
  
  // 系统通知
  'system:notification': {
    type: string;
    message: string;
    level: 'info' | 'warning' | 'error';
    timestamp: Timestamp;
  };
}

// API错误类型
export interface ApiError {
  code: string;
  message: string;
  details?: any;
  timestamp: Timestamp;
  requestId?: string;
  path?: string;
}

// 常见错误代码
export enum ErrorCodes {
  // 认证错误
  UNAUTHORIZED = 'UNAUTHORIZED',
  FORBIDDEN = 'FORBIDDEN',
  TOKEN_EXPIRED = 'TOKEN_EXPIRED',
  INVALID_CREDENTIALS = 'INVALID_CREDENTIALS',
  
  // 验证错误
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  INVALID_INPUT = 'INVALID_INPUT',
  MISSING_REQUIRED_FIELD = 'MISSING_REQUIRED_FIELD',
  
  // 资源错误
  NOT_FOUND = 'NOT_FOUND',
  ALREADY_EXISTS = 'ALREADY_EXISTS',
  CONFLICT = 'CONFLICT',
  
  // 业务错误
  INSUFFICIENT_BALANCE = 'INSUFFICIENT_BALANCE',
  POSITION_NOT_FOUND = 'POSITION_NOT_FOUND',
  SYNC_IN_PROGRESS = 'SYNC_IN_PROGRESS',
  KYC_REQUIRED = 'KYC_REQUIRED',
  REGION_RESTRICTED = 'REGION_RESTRICTED',
  
  // 系统错误
  INTERNAL_ERROR = 'INTERNAL_ERROR',
  SERVICE_UNAVAILABLE = 'SERVICE_UNAVAILABLE',
  RATE_LIMIT_EXCEEDED = 'RATE_LIMIT_EXCEEDED',
  MAINTENANCE_MODE = 'MAINTENANCE_MODE',
  
  // 外部服务错误
  EXTERNAL_SERVICE_ERROR = 'EXTERNAL_SERVICE_ERROR',
  BLOCKCHAIN_ERROR = 'BLOCKCHAIN_ERROR',
  PRICE_FEED_ERROR = 'PRICE_FEED_ERROR'
}

// API客户端配置
export interface ApiClientConfig {
  baseUrl: string;
  timeout: number;
  retries: number;
  retryDelay: number;
  headers?: Record<string, string>;
  
  // 认证配置
  auth?: {
    type: 'bearer' | 'api_key';
    token?: string;
    apiKey?: string;
  };
  
  // 缓存配置
  cache?: {
    enabled: boolean;
    ttl: number;
    maxSize: number;
  };
  
  // 拦截器
  interceptors?: {
    request?: (config: any) => any;
    response?: (response: any) => any;
    error?: (error: any) => any;
  };
}

// 批量操作类型
export interface BatchRequest<T = any> {
  operations: {
    method: 'GET' | 'POST' | 'PUT' | 'DELETE';
    url: string;
    data?: T;
    headers?: Record<string, string>;
  }[];
}

export interface BatchResponse<T = any> {
  results: {
    status: number;
    data?: T;
    error?: ApiError;
  }[];
}

// 实时订阅类型
export interface Subscription {
  id: UUID;
  type: 'price' | 'position' | 'alert' | 'news';
  filters: Record<string, any>;
  callback: (data: any) => void;
  createdAt: Timestamp;
}

// 导出报告类型
export interface ReportRequest {
  type: 'portfolio' | 'tax' | 'performance' | 'compliance';
  format: 'pdf' | 'csv' | 'xlsx';
  period: {
    from: Timestamp;
    to: Timestamp;
  };
  filters?: Record<string, any>;
  options?: {
    includeCharts: boolean;
    includeSummary: boolean;
    groupBy?: string;
  };
}

export interface ReportResponse {
  reportId: UUID;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  downloadUrl?: string;
  expiresAt?: Timestamp;
  createdAt: Timestamp;
}
