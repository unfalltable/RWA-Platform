import { UUID, Timestamp, Region, Currency } from './common';
import { AssetType } from './asset';

// 渠道类型
export enum ChannelType {
  EXCHANGE = 'exchange',        // 中心化交易所
  BROKER = 'broker',           // 券商
  DEX = 'dex',                 // 去中心化交易所
  ISSUER = 'issuer',           // 发行方直销
  BANK = 'bank',               // 银行
  PLATFORM = 'platform'        // 其他平台
}

// 渠道状态
export enum ChannelStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  MAINTENANCE = 'maintenance',
  SUSPENDED = 'suspended'
}

// 支付方式
export enum PaymentMethod {
  BANK_TRANSFER = 'bank_transfer',
  CREDIT_CARD = 'credit_card',
  DEBIT_CARD = 'debit_card',
  CRYPTO = 'crypto',
  PAYPAL = 'paypal',
  WIRE = 'wire'
}

// 渠道基础信息
export interface Channel {
  id: UUID;
  name: string;
  displayName: string;
  description: string;
  type: ChannelType;
  status: ChannelStatus;
  
  // 基础信息
  website: string;
  logo?: string;
  
  // 合规信息
  compliance: {
    licenses: {
      jurisdiction: Region;
      licenseType: string;
      licenseNumber: string;
      issuedDate: Timestamp;
      expiryDate?: Timestamp;
    }[];
    
    // 支持的地区
    supportedRegions: Region[];
    restrictedRegions: Region[];
    
    // KYC要求
    kycRequired: boolean;
    kycLevels: string[];
    
    // 合格投资者要求
    accreditedOnly: boolean;
    minimumNetWorth?: number;
  };
  
  // 支持的资产
  supportedAssets: {
    assetId: UUID;
    assetType: AssetType;
    tradingPairs: string[];
    minimumOrder?: number;
    maximumOrder?: number;
  }[];
  
  // 费用结构
  fees: {
    trading: {
      maker?: number;
      taker?: number;
      flat?: number;
    };
    deposit: {
      crypto?: number;
      fiat?: number;
      wire?: number;
    };
    withdrawal: {
      crypto?: number;
      fiat?: number;
      wire?: number;
    };
    management?: number;
  };
  
  // 支付方式
  paymentMethods: {
    method: PaymentMethod;
    currencies: Currency[];
    processingTime: string;
    limits: {
      min?: number;
      max?: number;
      daily?: number;
      monthly?: number;
    };
  }[];
  
  // 客服信息
  support: {
    email?: string;
    phone?: string;
    chat?: boolean;
    hours: string;
    languages: string[];
    responseTime: string;
  };
  
  // API信息
  api?: {
    hasReadOnlyApi: boolean;
    hasTradingApi: boolean;
    documentation?: string;
    rateLimits?: {
      requests: number;
      period: string;
    };
  };
  
  // 安全信息
  security: {
    insurance?: {
      coverage: number;
      provider: string;
    };
    custody: {
      type: 'self' | 'third_party';
      provider?: string;
      segregation: boolean;
    };
    audits: {
      auditor: string;
      reportDate: Timestamp;
      reportUrl?: string;
    }[];
  };
  
  // 时间戳
  createdAt: Timestamp;
  updatedAt: Timestamp;
}

// 渠道评分
export interface ChannelRating {
  channelId: UUID;
  
  // 综合评分
  overallScore: number;
  
  // 分项评分
  scores: {
    security: number;        // 安全性
    compliance: number;      // 合规性
    liquidity: number;       // 流动性
    fees: number;           // 费用合理性
    userExperience: number; // 用户体验
    support: number;        // 客服质量
    reputation: number;     // 声誉
  };
  
  // 用户评价
  userReviews: {
    totalReviews: number;
    averageRating: number;
    distribution: {
      [key: number]: number; // 1-5星分布
    };
  };
  
  // 风险事件
  riskEvents: {
    type: 'hack' | 'outage' | 'regulatory' | 'liquidity' | 'other';
    severity: 'low' | 'medium' | 'high' | 'critical';
    description: string;
    date: Timestamp;
    resolved: boolean;
    impact?: string;
  }[];
  
  // 更新时间
  updatedAt: Timestamp;
}

// 渠道匹配结果
export interface ChannelMatch {
  channelId: UUID;
  channel: Channel;
  
  // 匹配度
  matchScore: number;
  
  // 可用性
  availability: {
    available: boolean;
    reasons?: string[];
  };
  
  // 估算费用
  estimatedFees: {
    total: number;
    breakdown: {
      trading?: number;
      deposit?: number;
      withdrawal?: number;
    };
  };
  
  // 跳转信息
  redirect: {
    url: string;
    method: 'GET' | 'POST';
    parameters?: Record<string, string>;
    attribution?: {
      source: string;
      medium: string;
      campaign?: string;
    };
  };
  
  // 预计处理时间
  processingTime: {
    kyc?: string;
    deposit?: string;
    trade?: string;
    withdrawal?: string;
  };
}

// 渠道筛选条件
export interface ChannelFilter {
  types?: ChannelType[];
  regions?: Region[];
  assetTypes?: AssetType[];
  
  // 费用范围
  maxTradingFee?: number;
  maxWithdrawalFee?: number;
  
  // 合规要求
  kycRequired?: boolean;
  accreditedOnly?: boolean;
  
  // 支付方式
  paymentMethods?: PaymentMethod[];
  
  // 最低评分
  minRating?: number;
  
  // 其他要求
  hasInsurance?: boolean;
  hasApi?: boolean;
}
