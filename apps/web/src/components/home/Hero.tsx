import React from 'react';
import Link from 'next/link';
import { useTranslation } from 'next-i18next';
import { ArrowRight, TrendingUp, Shield, Globe, Zap } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { useAuth } from '@/contexts/AuthContext';

export function Hero() {
  const { t } = useTranslation('home');
  const { isAuthenticated } = useAuth();

  const features = [
    {
      icon: <TrendingUp className="w-6 h-6 text-brand-600" />,
      title: t('hero.features.aggregation.title'),
      description: t('hero.features.aggregation.description'),
    },
    {
      icon: <Shield className="w-6 h-6 text-brand-600" />,
      title: t('hero.features.compliance.title'),
      description: t('hero.features.compliance.description'),
    },
    {
      icon: <Globe className="w-6 h-6 text-brand-600" />,
      title: t('hero.features.global.title'),
      description: t('hero.features.global.description'),
    },
    {
      icon: <Zap className="w-6 h-6 text-brand-600" />,
      title: t('hero.features.realtime.title'),
      description: t('hero.features.realtime.description'),
    },
  ];

  const stats = [
    {
      value: '$2.5B+',
      label: t('hero.stats.totalValue'),
    },
    {
      value: '150+',
      label: t('hero.stats.assets'),
    },
    {
      value: '50+',
      label: t('hero.stats.channels'),
    },
    {
      value: '10K+',
      label: t('hero.stats.users'),
    },
  ];

  return (
    <section className="relative overflow-hidden bg-gradient-to-br from-brand-50 via-white to-blue-50">
      {/* 背景装饰 */}
      <div className="absolute inset-0">
        <div className="absolute top-0 left-0 w-96 h-96 bg-brand-200 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse-slow"></div>
        <div className="absolute top-0 right-0 w-96 h-96 bg-blue-200 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse-slow" style={{ animationDelay: '2s' }}></div>
        <div className="absolute bottom-0 left-1/2 w-96 h-96 bg-purple-200 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-pulse-slow" style={{ animationDelay: '4s' }}></div>
      </div>

      <div className="relative container-responsive py-20 lg:py-32">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* 左侧内容 */}
          <div className="text-center lg:text-left">
            <div className="inline-flex items-center px-4 py-2 bg-brand-100 text-brand-800 rounded-full text-sm font-medium mb-6">
              <span className="w-2 h-2 bg-brand-600 rounded-full mr-2 animate-pulse"></span>
              {t('hero.badge')}
            </div>

            <h1 className="text-4xl lg:text-6xl font-bold text-gray-900 mb-6 leading-tight">
              <span className="text-gradient">
                {t('hero.title.highlight')}
              </span>
              <br />
              {t('hero.title.main')}
            </h1>

            <p className="text-xl text-gray-600 mb-8 max-w-2xl">
              {t('hero.description')}
            </p>

            <div className="flex flex-col sm:flex-row gap-4 mb-12">
              {isAuthenticated ? (
                <Link href="/portfolio">
                  <Button size="lg" className="group">
                    {t('hero.cta.portfolio')}
                    <ArrowRight className="ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
                  </Button>
                </Link>
              ) : (
                <Link href="/register">
                  <Button size="lg" className="group">
                    {t('hero.cta.getStarted')}
                    <ArrowRight className="ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
                  </Button>
                </Link>
              )}
              
              <Link href="/assets">
                <Button variant="outline" size="lg">
                  {t('hero.cta.exploreAssets')}
                </Button>
              </Link>
            </div>

            {/* 统计数据 */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
              {stats.map((stat, index) => (
                <div key={index} className="text-center lg:text-left">
                  <div className="text-2xl lg:text-3xl font-bold text-gray-900">
                    {stat.value}
                  </div>
                  <div className="text-sm text-gray-600 mt-1">
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 右侧特性卡片 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            {features.map((feature, index) => (
              <Card
                key={index}
                className="p-6 hover:shadow-lg transition-all duration-300 hover:-translate-y-1"
                style={{
                  animationDelay: `${index * 0.1}s`,
                }}
              >
                <div className="flex items-start space-x-4">
                  <div className="flex-shrink-0">
                    <div className="w-12 h-12 bg-brand-100 rounded-lg flex items-center justify-center">
                      {feature.icon}
                    </div>
                  </div>
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                      {feature.title}
                    </h3>
                    <p className="text-gray-600 text-sm">
                      {feature.description}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      </div>

      {/* 底部波浪装饰 */}
      <div className="absolute bottom-0 left-0 right-0">
        <svg
          className="w-full h-20 text-white"
          viewBox="0 0 1200 120"
          preserveAspectRatio="none"
        >
          <path
            d="M0,0V46.29c47.79,22.2,103.59,32.17,158,28,70.36-5.37,136.33-33.31,206.8-37.5C438.64,32.43,512.34,53.67,583,72.05c69.27,18,138.3,24.88,209.4,13.08,36.15-6,69.85-17.84,104.45-29.34C989.49,25,1113-14.29,1200,52.47V0Z"
            opacity=".25"
            fill="currentColor"
          ></path>
          <path
            d="M0,0V15.81C13,36.92,27.64,56.86,47.69,72.05,99.41,111.27,165,111,224.58,91.58c31.15-10.15,60.09-26.07,89.67-39.8,40.92-19,84.73-46,130.83-49.67,36.26-2.85,70.9,9.42,98.6,31.56,31.77,25.39,62.32,62,103.63,73,40.44,10.79,81.35-6.69,119.13-24.28s75.16-39,116.92-43.05c59.73-5.85,113.28,22.88,168.9,38.84,30.2,8.66,59,6.17,87.09-7.5,22.43-10.89,48-26.93,60.65-49.24V0Z"
            opacity=".5"
            fill="currentColor"
          ></path>
          <path
            d="M0,0V5.63C149.93,59,314.09,71.32,475.83,42.57c43-7.64,84.23-20.12,127.61-26.46,59-8.63,112.48,12.24,165.56,35.4C827.93,77.22,886,95.24,951.2,90c86.53-7,172.46-45.71,248.8-84.81V0Z"
            fill="currentColor"
          ></path>
        </svg>
      </div>
    </section>
  );
}
