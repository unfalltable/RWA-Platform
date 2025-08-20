import { UUID, Timestamp, Region } from './common';
import { AssetType } from './asset';

// 风险类型
export enum RiskType {
  MARKET = 'market',           // 市场风险
  CREDIT = 'credit',           // 信用风险
  LIQUIDITY = 'liquidity',     // 流动性风险
  OPERATIONAL = 'operational', // 操作风险
  REGULATORY = 'regulatory',   // 监管风险
  TECHNICAL = 'technical',     // 技术风险
  CONCENTRATION = 'concentration', // 集中度风险
  COUNTERPARTY = 'counterparty'   // 对手方风险
}

// 风险等级
export enum RiskLevel {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  CRITICAL = 'critical'
}

// 预警状态
export enum AlertStatus {
  ACTIVE = 'active',
  ACKNOWLEDGED = 'acknowledged',
  RESOLVED = 'resolved',
  DISMISSED = 'dismissed'
}

// 风险评估
export interface RiskAssessment {
  id: UUID;
  entityId: UUID; // 可以是资产ID、用户ID、渠道ID等
  entityType: 'asset' | 'user' | 'channel' | 'portfolio';
  
  // 综合风险评分
  overallScore: number; // 0-100
  overallLevel: RiskLevel;
  
  // 分项风险评分
  riskScores: {
    [key in RiskType]?: {
      score: number;
      level: RiskLevel;
      weight: number;
      factors: string[];
    };
  };
  
  // 风险因子
  riskFactors: {
    factor: string;
    impact: 'positive' | 'negative';
    severity: number;
    description: string;
    source: string;
  }[];
  
  // 风险建议
  recommendations: {
    priority: 'high' | 'medium' | 'low';
    action: string;
    description: string;
    estimatedImpact?: number;
  }[];
  
  // 历史评分
  history: {
    date: Timestamp;
    score: number;
    level: RiskLevel;
    changeReason?: string;
  }[];
  
  // 评估信息
  assessedBy: 'system' | 'analyst' | 'model';
  assessedAt: Timestamp;
  validUntil?: Timestamp;
  
  // 元数据
  metadata: {
    model?: string;
    version?: string;
    confidence?: number;
    dataQuality?: number;
  };
}

// 风险预警
export interface RiskAlert {
  id: UUID;
  userId?: UUID;
  entityId: UUID;
  entityType: 'asset' | 'position' | 'portfolio' | 'channel';
  
  // 预警信息
  type: RiskType;
  level: RiskLevel;
  status: AlertStatus;
  
  // 预警内容
  title: string;
  message: string;
  description: string;
  
  // 触发条件
  trigger: {
    rule: string;
    threshold: number;
    currentValue: number;
    condition: 'above' | 'below' | 'equal' | 'change';
  };
  
  // 影响评估
  impact: {
    estimatedLoss?: number;
    affectedPositions?: number;
    affectedValue?: number;
    probability?: number;
  };
  
  // 建议行动
  actions: {
    priority: 'immediate' | 'urgent' | 'normal';
    action: string;
    description: string;
    deadline?: Timestamp;
  }[];
  
  // 相关数据
  relatedData: {
    [key: string]: any;
  };
  
  // 时间信息
  triggeredAt: Timestamp;
  acknowledgedAt?: Timestamp;
  resolvedAt?: Timestamp;
  expiresAt?: Timestamp;
  
  // 通知信息
  notifications: {
    email: boolean;
    sms: boolean;
    push: boolean;
    inApp: boolean;
    sentAt?: Timestamp;
  };
}

// 风险规则
export interface RiskRule {
  id: UUID;
  name: string;
  description: string;
  
  // 规则配置
  config: {
    entityType: 'asset' | 'position' | 'portfolio' | 'channel';
    riskType: RiskType;
    
    // 触发条件
    conditions: {
      metric: string;
      operator: 'gt' | 'lt' | 'eq' | 'gte' | 'lte' | 'change_gt' | 'change_lt';
      value: number;
      timeWindow?: number; // 时间窗口（分钟）
    }[];
    
    // 逻辑关系
    logic: 'and' | 'or';
    
    // 阈值配置
    thresholds: {
      [key in RiskLevel]?: number;
    };
  };
  
  // 规则状态
  status: 'active' | 'inactive' | 'testing';
  priority: number;
  
  // 适用范围
  scope: {
    assetTypes?: AssetType[];
    regions?: Region[];
    userTypes?: string[];
    minValue?: number;
    maxValue?: number;
  };
  
  // 执行配置
  execution: {
    frequency: number; // 检查频率（分钟）
    maxAlerts: number; // 最大预警数量
    cooldown: number;  // 冷却时间（分钟）
    autoResolve: boolean;
    autoResolveAfter?: number; // 自动解决时间（分钟）
  };
  
  // 通知配置
  notifications: {
    channels: ('email' | 'sms' | 'push' | 'webhook')[];
    recipients: string[];
    template?: string;
  };
  
  // 统计信息
  stats: {
    totalTriggers: number;
    falsePositives: number;
    truePositives: number;
    lastTriggered?: Timestamp;
    avgResolutionTime?: number;
  };
  
  // 时间信息
  createdAt: Timestamp;
  updatedAt: Timestamp;
  createdBy: UUID;
}

// 风险事件
export interface RiskEvent {
  id: UUID;
  
  // 事件信息
  type: RiskType;
  severity: RiskLevel;
  category: string;
  
  // 事件详情
  title: string;
  description: string;
  source: string;
  
  // 影响范围
  impact: {
    entityType: 'asset' | 'channel' | 'market' | 'protocol';
    entityIds: UUID[];
    estimatedLoss?: number;
    affectedUsers?: number;
    affectedValue?: number;
  };
  
  // 事件状态
  status: 'ongoing' | 'resolved' | 'investigating';
  
  // 时间线
  timeline: {
    timestamp: Timestamp;
    event: string;
    description: string;
    source?: string;
  }[];
  
  // 相关链接
  references: {
    type: 'news' | 'report' | 'announcement' | 'analysis';
    title: string;
    url: string;
    publishedAt?: Timestamp;
  }[];
  
  // 缓解措施
  mitigations: {
    action: string;
    description: string;
    implementedAt?: Timestamp;
    effectiveness?: 'high' | 'medium' | 'low';
  }[];
  
  // 时间信息
  occurredAt: Timestamp;
  reportedAt: Timestamp;
  resolvedAt?: Timestamp;
  
  // 元数据
  metadata: {
    confidence: number;
    dataQuality: number;
    sources: string[];
  };
}

// 合规检查
export interface ComplianceCheck {
  id: UUID;
  entityId: UUID;
  entityType: 'user' | 'transaction' | 'address' | 'asset';
  
  // 检查类型
  checkType: 'kyc' | 'aml' | 'sanctions' | 'pep' | 'address_screening' | 'transaction_monitoring';
  
  // 检查结果
  result: 'pass' | 'fail' | 'review' | 'pending';
  score?: number;
  
  // 检查详情
  details: {
    provider: string;
    checkId: string;
    flags: {
      type: string;
      severity: 'low' | 'medium' | 'high';
      description: string;
      confidence: number;
    }[];
    
    // 匹配信息
    matches?: {
      listType: string;
      matchType: 'exact' | 'fuzzy' | 'partial';
      confidence: number;
      details: any;
    }[];
  };
  
  // 风险评分
  riskScore: number;
  riskLevel: RiskLevel;
  
  // 建议行动
  recommendedAction: 'approve' | 'reject' | 'review' | 'monitor';
  
  // 审核信息
  review?: {
    reviewedBy: UUID;
    reviewedAt: Timestamp;
    decision: 'approve' | 'reject' | 'escalate';
    notes: string;
  };
  
  // 时间信息
  checkedAt: Timestamp;
  expiresAt?: Timestamp;
  
  // 元数据
  metadata: {
    version: string;
    cost?: number;
    processingTime?: number;
  };
}

// 风险配置
export interface RiskConfig {
  id: UUID;
  name: string;
  description: string;
  
  // 评分模型配置
  scoringModel: {
    weights: {
      [key in RiskType]?: number;
    };
    
    // 阈值配置
    thresholds: {
      [key in RiskLevel]: {
        min: number;
        max: number;
      };
    };
    
    // 衰减配置
    decay: {
      enabled: boolean;
      halfLife: number; // 半衰期（天）
      minWeight: number;
    };
  };
  
  // 预警配置
  alertConfig: {
    enabled: boolean;
    defaultLevel: RiskLevel;
    escalationRules: {
      level: RiskLevel;
      timeToEscalate: number; // 分钟
      escalateTo: RiskLevel;
    }[];
  };
  
  // 合规配置
  complianceConfig: {
    kycRequired: boolean;
    amlChecks: boolean;
    sanctionsScreening: boolean;
    pepScreening: boolean;
    addressScreening: boolean;
    
    // 检查频率
    recheckFrequency: {
      kyc: number; // 天
      sanctions: number;
      pep: number;
    };
  };
  
  // 适用范围
  scope: {
    regions: Region[];
    assetTypes: AssetType[];
    userTypes: string[];
  };
  
  // 版本信息
  version: string;
  effectiveFrom: Timestamp;
  effectiveTo?: Timestamp;
  
  createdAt: Timestamp;
  updatedAt: Timestamp;
  createdBy: UUID;
}
