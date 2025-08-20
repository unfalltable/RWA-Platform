import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, SelectQueryBuilder } from 'typeorm';
import { AssetEntity, AssetMetricEntity } from '../entities/asset.entity';
import { Asset, AssetMetrics, AssetRating } from './assets.types';
import { AssetFilterInput, PaginationInput, PaginatedResponse } from '../common/common.types';
import { RedisService } from '../redis/redis.service';

@Injectable()
export class AssetsService {
  private readonly logger = new Logger(AssetsService.name);

  constructor(
    @InjectRepository(AssetEntity)
    private readonly assetRepository: Repository<AssetEntity>,
    @InjectRepository(AssetMetricEntity)
    private readonly assetMetricRepository: Repository<AssetMetricEntity>,
    private readonly redisService: RedisService,
  ) {}

  async findAssets(
    filter?: AssetFilterInput,
    pagination?: PaginationInput,
  ): Promise<PaginatedResponse<Asset>> {
    const page = pagination?.page || 1;
    const limit = Math.min(pagination?.limit || 20, 100); // 限制最大100条
    const offset = (page - 1) * limit;

    let query = this.assetRepository.createQueryBuilder('asset');

    // 应用筛选条件
    if (filter) {
      query = this.applyFilters(query, filter);
    }

    // 应用排序
    const sortBy = pagination?.sortBy || 'createdAt';
    const sortOrder = pagination?.sortOrder === 'asc' ? 'ASC' : 'DESC';
    query = query.orderBy(`asset.${sortBy}`, sortOrder);

    // 获取总数
    const total = await query.getCount();

    // 应用分页
    const assets = await query.skip(offset).take(limit).getMany();

    return {
      data: assets.map(this.mapEntityToType),
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async findAssetById(id: string): Promise<Asset | null> {
    // 先尝试从缓存获取
    const cacheKey = `asset:${id}`;
    const cached = await this.redisService.get(cacheKey);
    if (cached) {
      return JSON.parse(cached);
    }

    const asset = await this.assetRepository.findOne({
      where: { id },
    });

    if (!asset) {
      return null;
    }

    const result = this.mapEntityToType(asset);

    // 缓存结果
    await this.redisService.setex(cacheKey, 300, JSON.stringify(result)); // 5分钟缓存

    return result;
  }

  async findAssetBySymbol(symbol: string): Promise<Asset | null> {
    const cacheKey = `asset:symbol:${symbol}`;
    const cached = await this.redisService.get(cacheKey);
    if (cached) {
      return JSON.parse(cached);
    }

    const asset = await this.assetRepository.findOne({
      where: { symbol: symbol.toUpperCase() },
    });

    if (!asset) {
      return null;
    }

    const result = this.mapEntityToType(asset);
    await this.redisService.setex(cacheKey, 300, JSON.stringify(result));

    return result;
  }

  async compareAssets(assetIds: string[]): Promise<Asset[]> {
    if (assetIds.length === 0) {
      return [];
    }

    const assets = await this.assetRepository.findByIds(assetIds);
    return assets.map(this.mapEntityToType);
  }

  async getAssetMetrics(assetId: string): Promise<AssetMetrics | null> {
    const cacheKey = `asset:metrics:${assetId}`;
    const cached = await this.redisService.get(cacheKey);
    if (cached) {
      return JSON.parse(cached);
    }

    // 获取最新的指标数据
    const latestMetric = await this.assetMetricRepository.findOne({
      where: { assetId },
      order: { timestamp: 'DESC' },
    });

    if (!latestMetric) {
      return null;
    }

    // 获取历史APY数据
    const apyHistory = await this.assetMetricRepository.find({
      where: { assetId },
      select: ['timestamp', 'apy'],
      order: { timestamp: 'DESC' },
      take: 30, // 最近30个数据点
    });

    // 获取波动率数据
    const volatilityData = await this.calculateVolatility(assetId);

    const metrics: AssetMetrics = {
      assetId,
      price: {
        value: latestMetric.price,
        currency: latestMetric.priceCurrency as any,
        timestamp: latestMetric.timestamp,
        source: 'aggregated',
      },
      priceChange: [
        { value: latestMetric.priceChange24h || 0, period: '24h' },
        { value: latestMetric.priceChange7d || 0, period: '7d' },
        { value: latestMetric.priceChange30d || 0, period: '30d' },
      ],
      apy: latestMetric.apy || 0,
      apyHistory: apyHistory.map(h => ({
        date: h.timestamp,
        value: h.apy || 0,
      })),
      volatility: volatilityData,
      volume24h: latestMetric.volume24h || 0,
      marketDepth: latestMetric.marketDepth || {},
      onChainMetrics: latestMetric.onChainMetrics || {},
      updatedAt: latestMetric.createdAt,
    };

    // 缓存结果
    await this.redisService.setex(cacheKey, 60, JSON.stringify(metrics)); // 1分钟缓存

    return metrics;
  }

  async getAssetRating(assetId: string): Promise<AssetRating | null> {
    const cacheKey = `asset:rating:${assetId}`;
    const cached = await this.redisService.get(cacheKey);
    if (cached) {
      return JSON.parse(cached);
    }

    const asset = await this.assetRepository.findOne({
      where: { id: assetId },
    });

    if (!asset || !asset.overallRating) {
      return null;
    }

    const rating: AssetRating = {
      assetId,
      overallScore: asset.overallRating,
      scores: asset.ratingScores || {
        transparency: 0,
        compliance: 0,
        liquidity: 0,
        stability: 0,
        yield: 0,
        security: 0,
      },
      rationale: {
        strengths: [],
        weaknesses: [],
        risks: [],
      },
      history: [],
      updatedAt: asset.updatedAt,
    };

    // 缓存结果
    await this.redisService.setex(cacheKey, 1800, JSON.stringify(rating)); // 30分钟缓存

    return rating;
  }

  async getRelatedAssets(assetId: string, assetType: string): Promise<Asset[]> {
    const assets = await this.assetRepository.find({
      where: { 
        type: assetType as any,
        id: Not(assetId),
      },
      take: 5,
      order: { marketCap: 'DESC' },
    });

    return assets.map(this.mapEntityToType);
  }

  async getMarketCapRank(assetId: string): Promise<number | null> {
    const result = await this.assetRepository
      .createQueryBuilder('asset')
      .select('COUNT(*) + 1', 'rank')
      .where('asset.marketCap > (SELECT marketCap FROM assets WHERE id = :assetId)', { assetId })
      .getRawOne();

    return result?.rank || null;
  }

  async getTopAssetsByType(type: string, limit: number): Promise<Asset[]> {
    const assets = await this.assetRepository.find({
      where: { type: type as any },
      order: { marketCap: 'DESC' },
      take: limit,
    });

    return assets.map(this.mapEntityToType);
  }

  async getTopAssetsByYield(limit: number): Promise<Asset[]> {
    const assets = await this.assetRepository.find({
      where: { apy: Not(IsNull()) },
      order: { apy: 'DESC' },
      take: limit,
    });

    return assets.map(this.mapEntityToType);
  }

  async getTopAssetsByMarketCap(limit: number): Promise<Asset[]> {
    const assets = await this.assetRepository.find({
      where: { marketCap: Not(IsNull()) },
      order: { marketCap: 'DESC' },
      take: limit,
    });

    return assets.map(this.mapEntityToType);
  }

  async searchAssets(query: string, limit: number): Promise<Asset[]> {
    const assets = await this.assetRepository
      .createQueryBuilder('asset')
      .where('asset.symbol ILIKE :query OR asset.name ILIKE :query', { 
        query: `%${query}%` 
      })
      .orderBy('asset.marketCap', 'DESC')
      .take(limit)
      .getMany();

    return assets.map(this.mapEntityToType);
  }

  async getTotalAssetsCount(): Promise<number> {
    return this.assetRepository.count();
  }

  async getTotalMarketCap(): Promise<number> {
    const result = await this.assetRepository
      .createQueryBuilder('asset')
      .select('SUM(asset.marketCap)', 'total')
      .where('asset.marketCap IS NOT NULL')
      .getRawOne();

    return result?.total || 0;
  }

  async getAssetTypeDistribution(): Promise<any[]> {
    return this.assetRepository
      .createQueryBuilder('asset')
      .select('asset.type', 'type')
      .addSelect('COUNT(*)', 'count')
      .addSelect('SUM(asset.marketCap)', 'totalMarketCap')
      .groupBy('asset.type')
      .orderBy('count', 'DESC')
      .getRawMany();
  }

  async getChainDistribution(): Promise<any[]> {
    // 这需要解析contracts JSON字段
    const result = await this.assetRepository
      .createQueryBuilder('asset')
      .select('jsonb_array_elements(asset.contracts)->\'chain\'', 'chain')
      .addSelect('COUNT(*)', 'count')
      .groupBy('chain')
      .orderBy('count', 'DESC')
      .getRawMany();

    return result;
  }

  private applyFilters(
    query: SelectQueryBuilder<AssetEntity>,
    filter: AssetFilterInput,
  ): SelectQueryBuilder<AssetEntity> {
    if (filter.types?.length) {
      query = query.andWhere('asset.type IN (:...types)', { types: filter.types });
    }

    if (filter.subTypes?.length) {
      query = query.andWhere('asset.subType IN (:...subTypes)', { subTypes: filter.subTypes });
    }

    if (filter.chains?.length) {
      // 搜索contracts JSON字段中的chain
      const chainConditions = filter.chains.map((chain, index) => 
        `asset.contracts @> '[{"chain": "${chain}"}]'`
      ).join(' OR ');
      query = query.andWhere(`(${chainConditions})`);
    }

    if (filter.apyRange) {
      if (filter.apyRange.min !== undefined) {
        query = query.andWhere('asset.apy >= :minApy', { minApy: filter.apyRange.min });
      }
      if (filter.apyRange.max !== undefined) {
        query = query.andWhere('asset.apy <= :maxApy', { maxApy: filter.apyRange.max });
      }
    }

    if (filter.kycRequired !== undefined) {
      query = query.andWhere('asset.compliance->>\'kycRequired\' = :kycRequired', { 
        kycRequired: filter.kycRequired.toString() 
      });
    }

    if (filter.search) {
      query = query.andWhere(
        '(asset.symbol ILIKE :search OR asset.name ILIKE :search OR asset.description ILIKE :search)',
        { search: `%${filter.search}%` }
      );
    }

    return query;
  }

  private async calculateVolatility(assetId: string): Promise<any[]> {
    // 计算不同时间段的波动率
    const periods = [
      { name: '7d', days: 7 },
      { name: '30d', days: 30 },
      { name: '90d', days: 90 },
    ];

    const volatilityData = [];

    for (const period of periods) {
      const startDate = new Date();
      startDate.setDate(startDate.getDate() - period.days);

      const prices = await this.assetMetricRepository.find({
        where: { 
          assetId,
          timestamp: MoreThan(startDate),
        },
        select: ['price'],
        order: { timestamp: 'ASC' },
      });

      if (prices.length > 1) {
        const returns = [];
        for (let i = 1; i < prices.length; i++) {
          const returnValue = (prices[i].price - prices[i-1].price) / prices[i-1].price;
          returns.push(returnValue);
        }

        // 计算标准差
        const mean = returns.reduce((sum, r) => sum + r, 0) / returns.length;
        const variance = returns.reduce((sum, r) => sum + Math.pow(r - mean, 2), 0) / returns.length;
        const volatility = Math.sqrt(variance) * Math.sqrt(365); // 年化波动率

        volatilityData.push({
          period: period.name,
          value: volatility,
        });
      }
    }

    return volatilityData;
  }

  private mapEntityToType(entity: AssetEntity): Asset {
    return {
      id: entity.id,
      symbol: entity.symbol,
      name: entity.name,
      description: entity.description,
      type: entity.type,
      subType: entity.subType,
      issuer: entity.issuer,
      contracts: entity.contracts,
      totalSupply: entity.totalSupply,
      marketCap: entity.marketCap,
      price: {
        value: entity.currentPrice || 0,
        currency: entity.priceCurrency as any,
        timestamp: entity.updatedAt,
      },
      apy: entity.apy,
      yieldInfo: entity.yieldInfo,
      fees: entity.fees,
      liquidity: entity.liquidity,
      compliance: entity.compliance,
      audits: entity.audits,
      custody: entity.custody,
      metadata: entity.metadata,
      createdAt: entity.createdAt,
      updatedAt: entity.updatedAt,
    };
  }
}

// 需要导入的类型
import { Not, IsNull, MoreThan } from 'typeorm';
