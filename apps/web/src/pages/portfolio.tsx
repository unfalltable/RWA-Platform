import React, { useState, useEffect } from 'react';
import type { NextPage, GetServerSideProps } from 'next';
import { useTranslation } from 'next-i18next';
import { serverSideTranslations } from 'next-i18next/serverSideTranslations';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  PieChart, 
  BarChart3,
  RefreshCw,
  Download,
  Settings,
  Eye,
  EyeOff,
  Calendar,
  Filter
} from 'lucide-react';

import { SEO } from '@/components/common/SEO';
import { Card, StatsCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { PortfolioOverview } from '@/components/portfolio/PortfolioOverview';
import { PositionsList } from '@/components/portfolio/PositionsList';
import { AllocationChart } from '@/components/portfolio/AllocationChart';
import { PerformanceChart } from '@/components/portfolio/PerformanceChart';
import { TransactionHistory } from '@/components/portfolio/TransactionHistory';
import { RiskAnalysis } from '@/components/portfolio/RiskAnalysis';
import { useAuth } from '@/contexts/AuthContext';
import { usePortfolio } from '@/hooks/usePortfolio';
import { formatCurrency, formatPercentage } from '@/lib/utils';
import { withAuth } from '@/lib/auth';

const PortfolioPage: NextPage = () => {
  const { t } = useTranslation('portfolio');
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('overview');
  const [showValues, setShowValues] = useState(true);
  const [timeRange, setTimeRange] = useState('1M');

  const {
    portfolio,
    loading,
    error,
    refreshPortfolio,
    syncPortfolio,
    isSyncing
  } = usePortfolio(user?.id);

  useEffect(() => {
    if (user?.id) {
      refreshPortfolio();
    }
  }, [user?.id]);

  const handleSync = async () => {
    try {
      await syncPortfolio();
    } catch (error) {
      console.error('Sync failed:', error);
    }
  };

  const handleExport = () => {
    // TODO: 实现导出功能
    console.log('Export portfolio data');
  };

  if (loading) {
    return (
      <div className="space-y-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-32 bg-gray-200 rounded-lg"></div>
            ))}
          </div>
          <div className="h-96 bg-gray-200 rounded-lg"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-96">
        <Card className="p-8 text-center max-w-md">
          <div className="text-red-600 mb-4">
            <TrendingDown className="w-12 h-12 mx-auto" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {t('error.title')}
          </h3>
          <p className="text-gray-600 mb-4">
            {t('error.description')}
          </p>
          <Button onClick={refreshPortfolio}>
            {t('error.retry')}
          </Button>
        </Card>
      </div>
    );
  }

  const portfolioStats = [
    {
      title: t('stats.totalValue'),
      value: showValues ? formatCurrency(portfolio?.totalValue || 0) : '••••••',
      icon: <DollarSign className="w-5 h-5 text-brand-600" />,
      trend: portfolio?.totalReturnPct >= 0 ? 'up' : 'down',
      change: { 
        value: Math.abs(portfolio?.totalReturnPct || 0), 
        period: t('stats.total') 
      },
    },
    {
      title: t('stats.dayChange'),
      value: showValues ? formatCurrency(portfolio?.dayChange || 0) : '••••••',
      icon: portfolio?.dayChange >= 0 ? 
        <TrendingUp className="w-5 h-5 text-green-600" /> : 
        <TrendingDown className="w-5 h-5 text-red-600" />,
      trend: portfolio?.dayChange >= 0 ? 'up' : 'down',
      change: { 
        value: Math.abs(portfolio?.dayChangePct || 0), 
        period: '24h' 
      },
    },
    {
      title: t('stats.totalReturn'),
      value: showValues ? formatCurrency(portfolio?.totalReturn || 0) : '••••••',
      icon: <BarChart3 className="w-5 h-5 text-brand-600" />,
      trend: portfolio?.totalReturn >= 0 ? 'up' : 'down',
      change: { 
        value: Math.abs(portfolio?.totalReturnPct || 0), 
        period: t('stats.total') 
      },
    },
    {
      title: t('stats.positions'),
      value: portfolio?.positions?.length || 0,
      icon: <PieChart className="w-5 h-5 text-brand-600" />,
      trend: 'neutral',
      change: { 
        value: 0, 
        period: '' 
      },
    },
  ];

  return (
    <>
      <SEO
        title={t('seo.title')}
        description={t('seo.description')}
        noIndex
      />

      <div className="space-y-8">
        {/* 页面头部 */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">
              {t('title')}
            </h1>
            <p className="mt-2 text-lg text-gray-600">
              {t('subtitle')}
            </p>
          </div>
          
          <div className="mt-4 lg:mt-0 flex items-center space-x-4">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowValues(!showValues)}
              leftIcon={showValues ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            >
              {showValues ? t('actions.hide') : t('actions.show')}
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={handleSync}
              loading={isSyncing}
              leftIcon={<RefreshCw className="w-4 h-4" />}
            >
              {t('actions.sync')}
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={handleExport}
              leftIcon={<Download className="w-4 h-4" />}
            >
              {t('actions.export')}
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              leftIcon={<Settings className="w-4 h-4" />}
            >
              {t('actions.settings')}
            </Button>
          </div>
        </div>

        {/* 投资组合统计 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {portfolioStats.map((stat, index) => (
            <StatsCard
              key={index}
              title={stat.title}
              value={stat.value}
              icon={stat.icon}
              trend={stat.trend}
              change={stat.change}
            />
          ))}
        </div>

        {/* 主要内容区域 */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between mb-6">
            <TabsList className="grid w-full lg:w-auto grid-cols-2 lg:grid-cols-6">
              <TabsTrigger value="overview">{t('tabs.overview')}</TabsTrigger>
              <TabsTrigger value="positions">{t('tabs.positions')}</TabsTrigger>
              <TabsTrigger value="allocation">{t('tabs.allocation')}</TabsTrigger>
              <TabsTrigger value="performance">{t('tabs.performance')}</TabsTrigger>
              <TabsTrigger value="transactions">{t('tabs.transactions')}</TabsTrigger>
              <TabsTrigger value="risk">{t('tabs.risk')}</TabsTrigger>
            </TabsList>

            {activeTab === 'performance' && (
              <div className="flex items-center space-x-2 mt-4 lg:mt-0">
                <Button
                  variant={timeRange === '1D' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('1D')}
                >
                  1D
                </Button>
                <Button
                  variant={timeRange === '1W' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('1W')}
                >
                  1W
                </Button>
                <Button
                  variant={timeRange === '1M' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('1M')}
                >
                  1M
                </Button>
                <Button
                  variant={timeRange === '3M' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('3M')}
                >
                  3M
                </Button>
                <Button
                  variant={timeRange === '1Y' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('1Y')}
                >
                  1Y
                </Button>
                <Button
                  variant={timeRange === 'ALL' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTimeRange('ALL')}
                >
                  ALL
                </Button>
              </div>
            )}
          </div>

          <TabsContent value="overview" className="space-y-6">
            <PortfolioOverview 
              portfolio={portfolio} 
              showValues={showValues}
            />
          </TabsContent>

          <TabsContent value="positions" className="space-y-6">
            <PositionsList 
              positions={portfolio?.positions || []} 
              showValues={showValues}
            />
          </TabsContent>

          <TabsContent value="allocation" className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Card className="p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  {t('allocation.byAssetType')}
                </h3>
                <AllocationChart 
                  data={portfolio?.allocation?.byAssetType || {}}
                  showValues={showValues}
                />
              </Card>
              
              <Card className="p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  {t('allocation.byChannel')}
                </h3>
                <AllocationChart 
                  data={portfolio?.allocation?.byChannel || {}}
                  showValues={showValues}
                />
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="performance" className="space-y-6">
            <Card className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">
                  {t('performance.title')}
                </h3>
                <Badge variant="outline">
                  {timeRange}
                </Badge>
              </div>
              <PerformanceChart 
                portfolio={portfolio}
                timeRange={timeRange}
                showValues={showValues}
              />
            </Card>

            {/* 性能指标 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <Card className="p-4">
                <div className="text-sm text-gray-600">{t('performance.sharpeRatio')}</div>
                <div className="text-2xl font-bold text-gray-900">
                  {showValues ? (portfolio?.performance?.sharpeRatio?.toFixed(2) || '--') : '••••'}
                </div>
              </Card>
              
              <Card className="p-4">
                <div className="text-sm text-gray-600">{t('performance.volatility')}</div>
                <div className="text-2xl font-bold text-gray-900">
                  {showValues ? formatPercentage(portfolio?.performance?.volatility || 0) : '••••'}
                </div>
              </Card>
              
              <Card className="p-4">
                <div className="text-sm text-gray-600">{t('performance.maxDrawdown')}</div>
                <div className="text-2xl font-bold text-red-600">
                  {showValues ? formatPercentage(portfolio?.performance?.maxDrawdown || 0) : '••••'}
                </div>
              </Card>
              
              <Card className="p-4">
                <div className="text-sm text-gray-600">{t('performance.cagr')}</div>
                <div className="text-2xl font-bold text-green-600">
                  {showValues ? formatPercentage(portfolio?.performance?.cagr || 0) : '••••'}
                </div>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="transactions" className="space-y-6">
            <TransactionHistory 
              userId={user?.id}
              showValues={showValues}
            />
          </TabsContent>

          <TabsContent value="risk" className="space-y-6">
            <RiskAnalysis 
              portfolio={portfolio}
              showValues={showValues}
            />
          </TabsContent>
        </Tabs>

        {/* 最后更新时间 */}
        {portfolio?.lastUpdated && (
          <div className="text-center text-sm text-gray-500">
            {t('lastUpdated')}: {new Date(portfolio.lastUpdated).toLocaleString()}
          </div>
        )}
      </div>
    </>
  );
};

export const getServerSideProps: GetServerSideProps = withAuth(async ({ locale }) => {
  return {
    props: {
      ...(await serverSideTranslations(locale ?? 'zh-CN', [
        'common',
        'navigation',
        'portfolio',
      ])),
    },
  };
});

export default PortfolioPage;
