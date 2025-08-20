import { UUID, Timestamp, Price, Currency, Chain, AddressInfo } from './common';
import { AssetType } from './asset';

// 持仓类型
export enum PositionType {
  SPOT = 'spot',           // 现货
  STAKING = 'staking',     // 质押
  LENDING = 'lending',     // 借贷
  VAULT = 'vault',         // 金库
  LP = 'liquidity_pool',   // 流动性池
  BOND = 'bond',           // 债券
  FUND = 'fund'            // 基金
}

// 持仓状态
export enum PositionStatus {
  ACTIVE = 'active',
  PENDING = 'pending',
  LOCKED = 'locked',
  MATURED = 'matured',
  REDEEMED = 'redeemed',
  ERROR = 'error'
}

// 持仓基础信息
export interface Position {
  id: UUID;
  userId: UUID;
  accountId: UUID;
  
  // 资产信息
  assetId: UUID;
  assetSymbol: string;
  assetType: AssetType;
  
  // 持仓类型和状态
  type: PositionType;
  status: PositionStatus;
  
  // 数量和价值
  quantity: number;
  decimals: number;
  
  // 成本信息
  costBasis: {
    totalCost: number;
    averagePrice: number;
    currency: Currency;
    fees: number;
  };
  
  // 当前价值
  currentValue: {
    price: Price;
    totalValue: number;
    unrealizedPnl: number;
    unrealizedPnlPercentage: number;
  };
  
  // 收益信息
  yield?: {
    apy: number;
    totalEarned: number;
    lastDistribution?: {
      amount: number;
      date: Timestamp;
      type: 'interest' | 'dividend' | 'reward';
    };
    nextDistribution?: {
      estimatedAmount: number;
      date: Timestamp;
    };
  };
  
  // 锁定信息
  lockup?: {
    isLocked: boolean;
    unlockDate?: Timestamp;
    earlyRedemptionPenalty?: number;
    minimumHoldPeriod?: number;
  };
  
  // 链上信息（适用于DeFi持仓）
  onChain?: {
    chain: Chain;
    contractAddress: string;
    tokenId?: string;
    transactionHash: string;
    blockNumber: number;
  };
  
  // 平台信息（适用于CeFi持仓）
  platform?: {
    name: string;
    accountId: string;
    productId?: string;
  };
  
  // 风险信息
  risk: {
    riskLevel: 'low' | 'medium' | 'high';
    liquidationPrice?: number;
    healthFactor?: number;
    warnings: string[];
  };
  
  // 时间信息
  openedAt: Timestamp;
  updatedAt: Timestamp;
  closedAt?: Timestamp;
}

// 持仓历史记录
export interface PositionHistory {
  id: UUID;
  positionId: UUID;
  
  // 操作类型
  action: 'open' | 'increase' | 'decrease' | 'close' | 'yield' | 'fee';
  
  // 操作详情
  details: {
    quantity?: number;
    price?: number;
    value?: number;
    fees?: number;
    description: string;
  };
  
  // 操作后状态
  afterAction: {
    totalQuantity: number;
    totalValue: number;
    averagePrice: number;
  };
  
  // 交易信息
  transaction?: {
    hash: string;
    blockNumber?: number;
    gasUsed?: number;
    gasFee?: number;
  };
  
  timestamp: Timestamp;
}

// 投资组合概览
export interface Portfolio {
  userId: UUID;
  
  // 总览
  summary: {
    totalValue: number;
    totalCost: number;
    totalPnl: number;
    totalPnlPercentage: number;
    totalYield: number;
    currency: Currency;
  };
  
  // 资产分布
  allocation: {
    byAssetType: {
      [key in AssetType]?: {
        value: number;
        percentage: number;
        count: number;
      };
    };
    byChain: {
      [key in Chain]?: {
        value: number;
        percentage: number;
        count: number;
      };
    };
    byPlatform: {
      [platform: string]: {
        value: number;
        percentage: number;
        count: number;
      };
    };
  };
  
  // 风险分析
  risk: {
    overallRiskLevel: 'low' | 'medium' | 'high';
    diversificationScore: number;
    concentrationRisk: {
      topAsset: number;
      top5Assets: number;
      top10Assets: number;
    };
    liquidityRisk: {
      liquid: number;
      illiquid: number;
      locked: number;
    };
  };
  
  // 收益分析
  performance: {
    totalReturn: number;
    totalReturnPercentage: number;
    annualizedReturn: number;
    sharpeRatio?: number;
    maxDrawdown?: number;
    volatility?: number;
    
    // 时间段收益
    returns: {
      '1d': number;
      '7d': number;
      '30d': number;
      '90d': number;
      '1y': number;
      'ytd': number;
      'all': number;
    };
  };
  
  // 现金流
  cashflow: {
    totalInflow: number;
    totalOutflow: number;
    netFlow: number;
    
    // 按类型分类
    byType: {
      deposits: number;
      withdrawals: number;
      yields: number;
      fees: number;
      trades: number;
    };
  };
  
  // 更新时间
  updatedAt: Timestamp;
}

// 现金流记录
export interface Cashflow {
  id: UUID;
  userId: UUID;
  positionId?: UUID;
  
  // 现金流类型
  type: 'deposit' | 'withdrawal' | 'yield' | 'fee' | 'trade' | 'transfer';
  direction: 'inflow' | 'outflow';
  
  // 金额信息
  amount: number;
  currency: Currency;
  usdValue?: number;
  
  // 资产信息
  assetId?: UUID;
  assetSymbol?: string;
  
  // 平台信息
  platform?: {
    name: string;
    accountId: string;
  };
  
  // 交易信息
  transaction?: {
    hash: string;
    blockNumber?: number;
    chain: Chain;
  };
  
  // 描述和标签
  description: string;
  tags: string[];
  category?: string;
  
  // 税务信息
  tax?: {
    isTaxable: boolean;
    taxCategory: string;
    costBasis?: number;
  };
  
  timestamp: Timestamp;
}

// 持仓同步状态
export interface PositionSync {
  userId: UUID;
  accountId: UUID;
  
  // 同步状态
  status: 'success' | 'error' | 'pending' | 'partial';
  
  // 同步信息
  lastSyncAt?: Timestamp;
  nextSyncAt?: Timestamp;
  syncFrequency: number; // 分钟
  
  // 错误信息
  error?: {
    code: string;
    message: string;
    retryCount: number;
    lastRetryAt?: Timestamp;
  };
  
  // 同步统计
  stats: {
    positionsFound: number;
    positionsUpdated: number;
    positionsAdded: number;
    positionsRemoved: number;
    cashflowsAdded: number;
  };
  
  // 数据源信息
  source: {
    type: 'api' | 'webhook' | 'manual';
    version?: string;
    rateLimit?: {
      remaining: number;
      resetAt: Timestamp;
    };
  };
}

// 持仓筛选条件
export interface PositionFilter {
  assetTypes?: AssetType[];
  positionTypes?: PositionType[];
  chains?: Chain[];
  platforms?: string[];
  
  // 价值范围
  valueRange?: {
    min?: number;
    max?: number;
  };
  
  // 收益范围
  yieldRange?: {
    min?: number;
    max?: number;
  };
  
  // 风险等级
  riskLevels?: ('low' | 'medium' | 'high')[];
  
  // 状态
  statuses?: PositionStatus[];
  
  // 时间范围
  dateRange?: {
    from?: Timestamp;
    to?: Timestamp;
  };
  
  // 搜索关键词
  search?: string;
}
