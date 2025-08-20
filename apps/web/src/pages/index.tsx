import type { NextPage, GetStaticProps } from 'next';
import Head from 'next/head';
import { useTranslation } from 'next-i18next';
import { serverSideTranslations } from 'next-i18next/serverSideTranslations';

import { Hero } from '@/components/home/Hero';
import { FeaturedAssets } from '@/components/home/FeaturedAssets';
import { MarketOverview } from '@/components/home/MarketOverview';
import { HowItWorks } from '@/components/home/HowItWorks';
import { TrustedBy } from '@/components/home/TrustedBy';
import { Newsletter } from '@/components/home/Newsletter';
import { SEO } from '@/components/common/SEO';

const HomePage: NextPage = () => {
  const { t } = useTranslation('common');

  return (
    <>
      <SEO
        title={t('seo.home.title')}
        description={t('seo.home.description')}
        keywords={t('seo.home.keywords')}
        canonical="/"
      />

      <main className="min-h-screen">
        {/* 英雄区域 */}
        <Hero />

        {/* 市场概览 */}
        <section className="py-16 bg-gray-50">
          <div className="container-responsive">
            <MarketOverview />
          </div>
        </section>

        {/* 精选资产 */}
        <section className="py-16">
          <div className="container-responsive">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-bold text-gray-900 mb-4">
                {t('home.featuredAssets.title')}
              </h2>
              <p className="text-lg text-gray-600 max-w-2xl mx-auto">
                {t('home.featuredAssets.description')}
              </p>
            </div>
            <FeaturedAssets />
          </div>
        </section>

        {/* 工作原理 */}
        <section className="py-16 bg-gray-50">
          <div className="container-responsive">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-bold text-gray-900 mb-4">
                {t('home.howItWorks.title')}
              </h2>
              <p className="text-lg text-gray-600 max-w-2xl mx-auto">
                {t('home.howItWorks.description')}
              </p>
            </div>
            <HowItWorks />
          </div>
        </section>

        {/* 合作伙伴 */}
        <section className="py-16">
          <div className="container-responsive">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-bold text-gray-900 mb-4">
                {t('home.trustedBy.title')}
              </h2>
              <p className="text-lg text-gray-600">
                {t('home.trustedBy.description')}
              </p>
            </div>
            <TrustedBy />
          </div>
        </section>

        {/* 订阅通讯 */}
        <section className="py-16 bg-brand-600">
          <div className="container-responsive">
            <Newsletter />
          </div>
        </section>
      </main>
    </>
  );
};

export const getStaticProps: GetStaticProps = async ({ locale }) => {
  return {
    props: {
      ...(await serverSideTranslations(locale ?? 'zh-CN', [
        'common',
        'navigation',
        'home',
      ])),
    },
    revalidate: 3600, // 1小时重新生成
  };
};

export default HomePage;
