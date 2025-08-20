import React, { useState, useEffect } from 'react';
import type { NextPage, GetStaticProps } from 'next';
import { useTranslation } from 'next-i18next';
import { serverSideTranslations } from 'next-i18next/serverSideTranslations';
import { 
  Search, 
  Filter, 
  Star, 
  Shield, 
  Globe, 
  CreditCard,
  ExternalLink,
  TrendingUp,
  Clock,
  CheckCircle
} from 'lucide-react';

import { SEO } from '@/components/common/SEO';
import { Card, StatsCard } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { ChannelCard } from '@/components/channels/ChannelCard';
import { ChannelFilters } from '@/components/channels/ChannelFilters';
import { ChannelComparison } from '@/components/channels/ChannelComparison';
import { useChannels } from '@/hooks/useChannels';
import { formatCurrency, formatPercentage } from '@/lib/utils';

const ChannelsPage: NextPage = () => {
  const { t } = useTranslation('channels');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedFilters, setSelectedFilters] = useState({});
  const [showFilters, setShowFilters] = useState(false);
  const [compareChannels, setCompareChannels] = useState<string[]>([]);
  const [showComparison, setShowComparison] = useState(false);

  const {
    channels,
    loading,
    error,
    pagination,
    fetchChannels,
    stats
  } = useChannels({
    search: searchQuery,
    filters: selectedFilters,
  });

  useEffect(() => {
    fetchChannels();
  }, [searchQuery, selectedFilters]);

  const handleSearch = (query: string) => {
    setSearchQuery(query);
  };

  const handleFilterChange = (filters: any) => {
    setSelectedFilters(filters);
  };

  const handleCompareToggle = (channelId: string) => {
    setCompareChannels(prev => {
      if (prev.includes(channelId)) {
        return prev.filter(id => id !== channelId);
      } else if (prev.length < 3) {
        return [...prev, channelId];
      }
      return prev;
    });
  };

  const handleShowComparison = () => {
    setShowComparison(true);
  };

  const marketStats = [
    {
      title: t('stats.totalChannels'),
      value: stats?.totalChannels || 0,
      icon: <Globe className="w-5 h-5 text-brand-600" />,
      trend: 'up',
      change: { value: 12, period: '30d' },
    },
    {
      title: t('stats.averageFee'),
      value: formatPercentage(stats?.averageFee || 0),
      icon: <CreditCard className="w-5 h-5 text-brand-600" />,
      trend: 'down',
      change: { value: -5, period: '30d' },
    },
    {
      title: t('stats.totalVolume'),
      value: formatCurrency(stats?.totalVolume || 0),
      icon: <TrendingUp className="w-5 h-5 text-brand-600" />,
      trend: 'up',
      change: { value: 23, period: '30d' },
    },
    {
      title: t('stats.activeChannels'),
      value: stats?.activeChannels || 0,
      icon: <CheckCircle className="w-5 h-5 text-brand-600" />,
      trend: 'up',
      change: { value: 8, period: '30d' },
    },
  ];

  return (
    <>
      <SEO
        title={t('seo.title')}
        description={t('seo.description')}
        keywords={t('seo.keywords')}
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
            {compareChannels.length > 0 && (
              <Button
                variant="outline"
                onClick={handleShowComparison}
                className="relative"
              >
                {t('compare')} ({compareChannels.length})
                {compareChannels.length > 0 && (
                  <span className="absolute -top-2 -right-2 w-5 h-5 bg-brand-600 text-white text-xs rounded-full flex items-center justify-center">
                    {compareChannels.length}
                  </span>
                )}
              </Button>
            )}
          </div>
        </div>

        {/* 市场统计 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {marketStats.map((stat, index) => (
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

        {/* 搜索和筛选 */}
        <Card className="p-6">
          <div className="flex flex-col lg:flex-row lg:items-center space-y-4 lg:space-y-0 lg:space-x-4">
            {/* 搜索框 */}
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                <Input
                  type="text"
                  placeholder={t('searchPlaceholder')}
                  value={searchQuery}
                  onChange={(e) => handleSearch(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            {/* 筛选按钮 */}
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
              leftIcon={<Filter className="w-4 h-4" />}
            >
              {t('filters')}
            </Button>
          </div>

          {/* 筛选面板 */}
          {showFilters && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <ChannelFilters
                filters={selectedFilters}
                onChange={handleFilterChange}
              />
            </div>
          )}
        </Card>

        {/* 渠道列表 */}
        <div className="space-y-6">
          {loading ? (
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
              {[...Array(6)].map((_, index) => (
                <Card key={index} className="p-6 animate-pulse">
                  <div className="flex items-center space-x-4">
                    <div className="w-12 h-12 bg-gray-200 rounded-lg"></div>
                    <div className="flex-1">
                      <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                      <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                    </div>
                  </div>
                  <div className="mt-4 space-y-2">
                    <div className="h-3 bg-gray-200 rounded"></div>
                    <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                  </div>
                </Card>
              ))}
            </div>
          ) : error ? (
            <Card className="p-8 text-center">
              <div className="text-red-600 mb-4">
                <Shield className="w-12 h-12 mx-auto" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {t('error.title')}
              </h3>
              <p className="text-gray-600 mb-4">
                {t('error.description')}
              </p>
              <Button onClick={() => fetchChannels()}>
                {t('error.retry')}
              </Button>
            </Card>
          ) : channels.length === 0 ? (
            <Card className="p-8 text-center">
              <div className="text-gray-400 mb-4">
                <Search className="w-12 h-12 mx-auto" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {t('noResults.title')}
              </h3>
              <p className="text-gray-600">
                {t('noResults.description')}
              </p>
            </Card>
          ) : (
            <>
              <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
                {channels.map((channel) => (
                  <ChannelCard
                    key={channel.id}
                    channel={channel}
                    isSelected={compareChannels.includes(channel.id)}
                    onCompareToggle={() => handleCompareToggle(channel.id)}
                    showCompareButton={compareChannels.length < 3 || compareChannels.includes(channel.id)}
                  />
                ))}
              </div>

              {/* 分页 */}
              {pagination && pagination.totalPages > 1 && (
                <div className="flex justify-center">
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      disabled={pagination.page === 1}
                      onClick={() => fetchChannels(pagination.page - 1)}
                    >
                      {t('pagination.previous')}
                    </Button>
                    
                    <span className="text-sm text-gray-600">
                      {t('pagination.info', {
                        current: pagination.page,
                        total: pagination.totalPages,
                      })}
                    </span>
                    
                    <Button
                      variant="outline"
                      disabled={pagination.page === pagination.totalPages}
                      onClick={() => fetchChannels(pagination.page + 1)}
                    >
                      {t('pagination.next')}
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* 渠道对比模态框 */}
      {showComparison && (
        <ChannelComparison
          channelIds={compareChannels}
          onClose={() => setShowComparison(false)}
        />
      )}
    </>
  );
};

export const getStaticProps: GetStaticProps = async ({ locale }) => {
  return {
    props: {
      ...(await serverSideTranslations(locale ?? 'zh-CN', [
        'common',
        'navigation',
        'channels',
      ])),
    },
    revalidate: 3600, // 1小时重新生成
  };
};

export default ChannelsPage;
