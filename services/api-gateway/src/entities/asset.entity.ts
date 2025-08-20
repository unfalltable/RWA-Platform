import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  Index,
  OneToMany,
} from 'typeorm';
import { AssetType, AssetSubType, Region, Chain } from '@rwa-platform/types';

@Entity('assets')
@Index(['type', 'status'])
@Index(['symbol'])
@Index(['createdAt'])
export class AssetEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  symbol: string;

  @Column()
  name: string;

  @Column('text')
  description: string;

  @Column({
    type: 'enum',
    enum: AssetType,
  })
  type: AssetType;

  @Column({
    type: 'enum',
    enum: AssetSubType,
    nullable: true,
  })
  subType?: AssetSubType;

  @Column({ default: 'active' })
  status: 'active' | 'inactive' | 'delisted';

  // 发行信息
  @Column('jsonb')
  issuer: {
    name: string;
    website?: string;
    jurisdiction: Region;
    licenses: string[];
  };

  // 合约信息
  @Column('jsonb')
  contracts: {
    chain: Chain;
    address: string;
    decimals: number;
    standard: string;
  }[];

  // 基础指标
  @Column('decimal', { precision: 20, scale: 8, nullable: true })
  totalSupply?: number;

  @Column('decimal', { precision: 20, scale: 2, nullable: true })
  marketCap?: number;

  @Column('decimal', { precision: 10, scale: 4, nullable: true })
  currentPrice?: number;

  @Column({ default: 'USD' })
  priceCurrency: string;

  // 收益信息
  @Column('decimal', { precision: 8, scale: 4, nullable: true })
  apy?: number;

  @Column('jsonb', { nullable: true })
  yieldInfo?: {
    current: number;
    historical: {
      period: string;
      value: number;
    }[];
  };

  // 费用结构
  @Column('jsonb')
  fees: {
    management?: number;
    performance?: number;
    redemption?: number;
    entry?: number;
  };

  // 流动性信息
  @Column('jsonb')
  liquidity: {
    dailyVolume?: number;
    marketDepth?: number;
    redemptionPeriod?: string;
    minimumInvestment?: number;
  };

  // 合规信息
  @Column('jsonb')
  compliance: {
    kycRequired: boolean;
    accreditedOnly: boolean;
    restrictedRegions: Region[];
    regulatoryFramework: string[];
  };

  // 审计信息
  @Column('jsonb')
  audits: {
    auditor: string;
    reportDate: number;
    reportUrl?: string;
    scope: string;
  }[];

  // 托管信息
  @Column('jsonb')
  custody: {
    custodian: string;
    insuranceCoverage?: number;
    segregation: boolean;
  }[];

  // 元数据
  @Column('jsonb')
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

  // 评分信息
  @Column('decimal', { precision: 5, scale: 2, nullable: true })
  overallRating?: number;

  @Column('jsonb', { nullable: true })
  ratingScores?: {
    transparency: number;
    compliance: number;
    liquidity: number;
    stability: number;
    yield: number;
    security: number;
  };

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;

  // 关联关系
  @OneToMany(() => AssetMetricEntity, (metric) => metric.asset)
  metrics: AssetMetricEntity[];

  @OneToMany(() => PositionEntity, (position) => position.asset)
  positions: PositionEntity[];
}

@Entity('asset_metrics')
@Index(['assetId', 'timestamp'])
@Index(['timestamp'])
export class AssetMetricEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column('uuid')
  assetId: string;

  // 价格指标
  @Column('decimal', { precision: 20, scale: 8 })
  price: number;

  @Column({ default: 'USD' })
  priceCurrency: string;

  @Column('decimal', { precision: 10, scale: 4, nullable: true })
  priceChange24h?: number;

  @Column('decimal', { precision: 10, scale: 4, nullable: true })
  priceChange7d?: number;

  @Column('decimal', { precision: 10, scale: 4, nullable: true })
  priceChange30d?: number;

  // 收益指标
  @Column('decimal', { precision: 8, scale: 4, nullable: true })
  apy?: number;

  @Column('decimal', { precision: 8, scale: 4, nullable: true })
  volatility?: number;

  // 流动性指标
  @Column('decimal', { precision: 20, scale: 2, nullable: true })
  volume24h?: number;

  @Column('decimal', { precision: 20, scale: 2, nullable: true })
  marketDepth?: number;

  // 链上指标
  @Column('jsonb', { nullable: true })
  onChainMetrics?: {
    holders: number;
    transactions24h: number;
    activeAddresses: number;
    concentration: {
      top10: number;
      top100: number;
    };
  };

  @Column('timestamp')
  timestamp: Date;

  @CreateDateColumn()
  createdAt: Date;

  // 关联关系
  @ManyToOne(() => AssetEntity, (asset) => asset.metrics)
  asset: AssetEntity;
}

// 导入其他实体的引用
import { PositionEntity } from './position.entity';
import { ManyToOne } from 'typeorm';
