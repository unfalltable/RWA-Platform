import { UUID, Timestamp, Price, PercentageChange, Chain, Currency, Region } from './common';

// 资产类型
export enum AssetType {
  STABLECOIN = 'stablecoin',
  TREASURY = 'treasury',
  MONEY_MARKET = 'money_market',
  GOLD = 'gold',
  REAL_ESTATE = 'real_estate',
  COMMODITY = 'commodity',
  BOND = 'bond'
}

// 资产子类型
export enum AssetSubType {
  // 稳定币
  FIAT_BACKED = 'fiat_backed',
  CRYPTO_BACKED = 'crypto_backed',
  ALGORITHMIC = 'algorithmic',
  
  // 国债
  SHORT_TERM = 'short_term',
  MEDIUM_TERM = 'medium_term',
  LONG_TERM = 'long_term',
  
  // 房地产
  RESIDENTIAL = 'residential',
  COMMERCIAL = 'commercial',
  REIT = 'reit'
}

// 资产基础信息
export interface Asset {
  id: UUID;
  symbol: string;
  name: string;
  description: string;
  type: AssetType;
  subType?: AssetSubType;
  
  // 发行信息
  issuer: {
    name: string;
    website?: string;
    jurisdiction: Region;
    licenses: string[];
  };
  
  // 合约信息
  contracts: {
    chain: Chain;
    address: string;
    decimals: number;
    standard: string; // ERC-20, ERC-4626, etc.
  }[];
  
  // 基础指标
  totalSupply?: number;
  marketCap?: number;
  price: Price;
  
  // 收益信息
  apy?: number;
  yield?: {
    current: number;
    historical: {
      period: string;
      value: number;
    }[];
  };
  
  // 费用结构
  fees: {
    management?: number;
    performance?: number;
    redemption?: number;
    entry?: number;
  };
  
  // 流动性信息
  liquidity: {
    dailyVolume?: number;
    marketDepth?: number;
    redemptionPeriod?: string;
    minimumInvestment?: number;
  };
  
  // 合规信息
  compliance: {
    kycRequired: boolean;
    accreditedOnly: boolean;
    restrictedRegions: Region[];
    regulatoryFramework: string[];
  };
  
  // 审计信息
  audits: {
    auditor: string;
    reportDate: Timestamp;
    reportUrl?: string;
    scope: string;
  }[];
  
  // 托管信息
  custody: {
    custodian: string;
    insuranceCoverage?: number;
    segregation: boolean;
  }[];
  
  // 元数据
  metadata: {
    logo?: string;
    website?: string;
    whitepaper?: string;
    documentation?: string;
    socialMedia?: {
      twitter?: string;
      telegram?: string;
      discord?: string;
    };
  };
  
  // 时间戳
  createdAt: Timestamp;
  updatedAt: Timestamp;
}

// 资产指标
export interface AssetMetrics {
  assetId: UUID;
  
  // 价格指标
  price: Price;
  priceChange: PercentageChange[];
  
  // 收益指标
  apy: number;
  apyHistory: {
    date: Timestamp;
    value: number;
  }[];
  
  // 风险指标
  volatility: {
    period: string;
    value: number;
  }[];
  
  // 流动性指标
  volume24h: number;
  marketDepth: {
    bids: number;
    asks: number;
  };
  
  // 链上指标（适用于代币化资产）
  onChainMetrics?: {
    holders: number;
    transactions24h: number;
    activeAddresses: number;
    concentration: {
      top10: number;
      top100: number;
    };
  };
  
  // 更新时间
  updatedAt: Timestamp;
}

// 资产评分
export interface AssetRating {
  assetId: UUID;
  
  // 综合评分
  overallScore: number;
  
  // 分项评分
  scores: {
    transparency: number;    // 透明度
    compliance: number;      // 合规性
    liquidity: number;       // 流动性
    stability: number;       // 稳定性
    yield: number;          // 收益性
    security: number;       // 安全性
  };
  
  // 评分说明
  rationale: {
    strengths: string[];
    weaknesses: string[];
    risks: string[];
  };
  
  // 评分历史
  history: {
    date: Timestamp;
    score: number;
    reason?: string;
  }[];
  
  // 更新时间
  updatedAt: Timestamp;
}

// 资产筛选条件
export interface AssetFilter {
  types?: AssetType[];
  subTypes?: AssetSubType[];
  chains?: Chain[];
  regions?: Region[];
  
  // 收益范围
  apyRange?: {
    min?: number;
    max?: number;
  };
  
  // 投资门槛
  minimumInvestment?: {
    min?: number;
    max?: number;
  };
  
  // 合规要求
  kycRequired?: boolean;
  accreditedOnly?: boolean;
  
  // 评分范围
  ratingRange?: {
    min?: number;
    max?: number;
  };
  
  // 搜索关键词
  search?: string;
}
