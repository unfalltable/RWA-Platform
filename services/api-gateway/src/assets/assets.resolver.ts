import { Resolver, Query, Args, Context, ResolveField, Parent } from '@nestjs/graphql';
import { UseGuards, Logger } from '@nestjs/common';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';
import { AssetsService } from './assets.service';
import { Asset, AssetConnection, AssetMetrics, AssetRating } from './assets.types';
import { AssetFilterInput, PaginationInput } from '../common/common.types';

@Resolver(() => Asset)
export class AssetsResolver {
  private readonly logger = new Logger(AssetsResolver.name);

  constructor(private readonly assetsService: AssetsService) {}

  @Query(() => AssetConnection)
  async assets(
    @Args('filter', { nullable: true }) filter?: AssetFilterInput,
    @Args('pagination', { nullable: true }) pagination?: PaginationInput,
    @Context() context?: any,
  ): Promise<AssetConnection> {
    this.logger.debug('Fetching assets with filter:', filter);
    
    const result = await this.assetsService.findAssets(filter, pagination);
    
    return {
      edges: result.data.map((asset, index) => ({
        node: asset,
        cursor: Buffer.from(`${result.page}-${index}`).toString('base64'),
      })),
      pageInfo: {
        hasNextPage: result.page < result.totalPages,
        hasPreviousPage: result.page > 1,
        startCursor: result.data.length > 0 ? Buffer.from(`${result.page}-0`).toString('base64') : null,
        endCursor: result.data.length > 0 ? Buffer.from(`${result.page}-${result.data.length - 1}`).toString('base64') : null,
      },
      totalCount: result.total,
    };
  }

  @Query(() => Asset, { nullable: true })
  async asset(
    @Args('id') id: string,
    @Context() context?: any,
  ): Promise<Asset | null> {
    this.logger.debug(`Fetching asset with id: ${id}`);
    return this.assetsService.findAssetById(id);
  }

  @Query(() => Asset, { nullable: true })
  async assetBySymbol(
    @Args('symbol') symbol: string,
    @Context() context?: any,
  ): Promise<Asset | null> {
    this.logger.debug(`Fetching asset with symbol: ${symbol}`);
    return this.assetsService.findAssetBySymbol(symbol);
  }

  @Query(() => [Asset])
  async compareAssets(
    @Args('assetIds', { type: () => [String] }) assetIds: string[],
    @Context() context?: any,
  ): Promise<Asset[]> {
    this.logger.debug(`Comparing assets: ${assetIds.join(', ')}`);
    return this.assetsService.compareAssets(assetIds);
  }

  // Field resolvers for lazy loading related data
  @ResolveField(() => AssetMetrics, { nullable: true })
  async metrics(@Parent() asset: Asset): Promise<AssetMetrics | null> {
    return this.assetsService.getAssetMetrics(asset.id);
  }

  @ResolveField(() => AssetRating, { nullable: true })
  async rating(@Parent() asset: Asset): Promise<AssetRating | null> {
    return this.assetsService.getAssetRating(asset.id);
  }

  @ResolveField(() => [Asset])
  async relatedAssets(@Parent() asset: Asset): Promise<Asset[]> {
    return this.assetsService.getRelatedAssets(asset.id, asset.type);
  }

  // 计算字段
  @ResolveField(() => Number, { nullable: true })
  async currentYield(@Parent() asset: Asset): Promise<number | null> {
    const metrics = await this.assetsService.getAssetMetrics(asset.id);
    return metrics?.apy || asset.apy || null;
  }

  @ResolveField(() => String)
  async riskLevel(@Parent() asset: Asset): Promise<string> {
    const rating = await this.assetsService.getAssetRating(asset.id);
    if (!rating) return 'UNKNOWN';
    
    if (rating.overallScore >= 80) return 'LOW';
    if (rating.overallScore >= 60) return 'MEDIUM';
    if (rating.overallScore >= 40) return 'HIGH';
    return 'CRITICAL';
  }

  @ResolveField(() => Boolean)
  async isActive(@Parent() asset: Asset): Promise<boolean> {
    // 检查资产是否活跃（有最近的价格数据）
    const metrics = await this.assetsService.getAssetMetrics(asset.id);
    if (!metrics) return false;
    
    const now = new Date();
    const lastUpdate = new Date(metrics.updatedAt);
    const hoursSinceUpdate = (now.getTime() - lastUpdate.getTime()) / (1000 * 60 * 60);
    
    return hoursSinceUpdate < 24; // 24小时内有更新认为是活跃的
  }

  @ResolveField(() => Number, { nullable: true })
  async marketCapRank(@Parent() asset: Asset): Promise<number | null> {
    return this.assetsService.getMarketCapRank(asset.id);
  }

  @ResolveField(() => String, { nullable: true })
  async priceChangeStatus(@Parent() asset: Asset): Promise<string | null> {
    const metrics = await this.assetsService.getAssetMetrics(asset.id);
    if (!metrics || !metrics.priceChange.length) return null;
    
    const change24h = metrics.priceChange.find(pc => pc.period === '24h');
    if (!change24h) return null;
    
    if (change24h.value > 5) return 'STRONG_UP';
    if (change24h.value > 1) return 'UP';
    if (change24h.value > -1) return 'STABLE';
    if (change24h.value > -5) return 'DOWN';
    return 'STRONG_DOWN';
  }

  // 聚合查询
  @Query(() => [Asset])
  async topAssetsByType(
    @Args('type') type: string,
    @Args('limit', { defaultValue: 10 }) limit: number,
  ): Promise<Asset[]> {
    return this.assetsService.getTopAssetsByType(type, limit);
  }

  @Query(() => [Asset])
  async topAssetsByYield(
    @Args('limit', { defaultValue: 10 }) limit: number,
  ): Promise<Asset[]> {
    return this.assetsService.getTopAssetsByYield(limit);
  }

  @Query(() => [Asset])
  async topAssetsByMarketCap(
    @Args('limit', { defaultValue: 10 }) limit: number,
  ): Promise<Asset[]> {
    return this.assetsService.getTopAssetsByMarketCap(limit);
  }

  @Query(() => [Asset])
  async searchAssets(
    @Args('query') query: string,
    @Args('limit', { defaultValue: 20 }) limit: number,
  ): Promise<Asset[]> {
    return this.assetsService.searchAssets(query, limit);
  }

  // 统计查询
  @Query(() => Number)
  async totalAssetsCount(): Promise<number> {
    return this.assetsService.getTotalAssetsCount();
  }

  @Query(() => Number)
  async totalMarketCap(): Promise<number> {
    return this.assetsService.getTotalMarketCap();
  }

  @Query(() => [Object])
  async assetTypeDistribution(): Promise<any[]> {
    return this.assetsService.getAssetTypeDistribution();
  }

  @Query(() => [Object])
  async chainDistribution(): Promise<any[]> {
    return this.assetsService.getChainDistribution();
  }
}
