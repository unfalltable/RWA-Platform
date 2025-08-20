import { Html, Head, Main, NextScript } from 'next/document';

export default function Document() {
  return (
    <Html lang="zh-CN">
      <Head>
        {/* 预连接到外部域名 */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
        
        {/* 网站图标 */}
        <link rel="icon" href="/favicon.ico" />
        <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
        <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
        <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/site.webmanifest" />
        
        {/* 主题颜色 */}
        <meta name="theme-color" content="#0ea5e9" />
        <meta name="msapplication-TileColor" content="#0ea5e9" />
        
        {/* SEO元标签 */}
        <meta name="robots" content="index,follow" />
        <meta name="googlebot" content="index,follow" />
        
        {/* Open Graph */}
        <meta property="og:type" content="website" />
        <meta property="og:site_name" content="RWA Platform" />
        <meta property="og:locale" content="zh_CN" />
        
        {/* Twitter Card */}
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:site" content="@rwaplatform" />
        
        {/* 安全策略 */}
        <meta httpEquiv="X-Content-Type-Options" content="nosniff" />
        <meta httpEquiv="X-Frame-Options" content="DENY" />
        <meta httpEquiv="X-XSS-Protection" content="1; mode=block" />
        <meta httpEquiv="Referrer-Policy" content="strict-origin-when-cross-origin" />
        
        {/* 预加载关键资源 */}
        <link
          rel="preload"
          href="/fonts/inter-var.woff2"
          as="font"
          type="font/woff2"
          crossOrigin=""
        />
        
        {/* 样式预加载 */}
        <link rel="preload" href="/styles/globals.css" as="style" />
        
        {/* DNS预解析 */}
        <link rel="dns-prefetch" href="//api.rwa-platform.com" />
        <link rel="dns-prefetch" href="//cdn.rwa-platform.com" />
        
        {/* 结构化数据 */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              '@context': 'https://schema.org',
              '@type': 'WebApplication',
              name: 'RWA Platform',
              description: '稳定资产聚合与撮合平台',
              url: 'https://rwa-platform.com',
              applicationCategory: 'FinanceApplication',
              operatingSystem: 'Web',
              offers: {
                '@type': 'Offer',
                price: '0',
                priceCurrency: 'USD',
              },
              author: {
                '@type': 'Organization',
                name: 'RWA Platform Team',
              },
            }),
          }}
        />
      </Head>
      <body className="antialiased">
        {/* 无JavaScript时的提示 */}
        <noscript>
          <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: '#f3f4f6',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 9999,
            fontSize: '18px',
            textAlign: 'center',
            padding: '20px',
          }}>
            <div>
              <h1>需要启用JavaScript</h1>
              <p>请在浏览器设置中启用JavaScript以使用RWA Platform。</p>
            </div>
          </div>
        </noscript>
        
        <Main />
        <NextScript />
        
        {/* 性能监控脚本 */}
        {process.env.NODE_ENV === 'production' && (
          <script
            dangerouslySetInnerHTML={{
              __html: `
                // 页面加载性能监控
                window.addEventListener('load', function() {
                  if ('performance' in window) {
                    const perfData = performance.getEntriesByType('navigation')[0];
                    if (perfData) {
                      // 发送性能数据到分析服务
                      console.log('Page load time:', perfData.loadEventEnd - perfData.fetchStart);
                    }
                  }
                });
                
                // 错误监控
                window.addEventListener('error', function(e) {
                  console.error('Global error:', e.error);
                  // 发送错误信息到监控服务
                });
                
                window.addEventListener('unhandledrejection', function(e) {
                  console.error('Unhandled promise rejection:', e.reason);
                  // 发送错误信息到监控服务
                });
              `,
            }}
          />
        )}
      </body>
    </Html>
  );
}
