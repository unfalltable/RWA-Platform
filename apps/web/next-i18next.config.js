module.exports = {
  i18n: {
    defaultLocale: 'zh-CN',
    locales: ['zh-CN', 'zh-TW', 'en', 'ja'],
    localeDetection: true,
  },
  fallbackLng: {
    'zh-HK': ['zh-TW', 'zh-CN'],
    'zh-SG': ['zh-CN'],
    default: ['zh-CN'],
  },
  nonExplicitSupportedLngs: true,
  cleanCode: true,
  returnNull: false,
  returnEmptyString: false,
  returnObjects: false,
  joinArrays: false,
  
  // 命名空间
  ns: [
    'common',
    'navigation',
    'assets',
    'portfolio',
    'channels',
    'auth',
    'risk',
    'errors'
  ],
  defaultNS: 'common',
  
  // 插值配置
  interpolation: {
    escapeValue: false,
    formatSeparator: ',',
    format: function(value, format, lng) {
      if (format === 'uppercase') return value.toUpperCase();
      if (format === 'lowercase') return value.toLowerCase();
      if (format === 'currency') {
        return new Intl.NumberFormat(lng, {
          style: 'currency',
          currency: 'USD'
        }).format(value);
      }
      if (format === 'percentage') {
        return new Intl.NumberFormat(lng, {
          style: 'percent',
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        }).format(value / 100);
      }
      return value;
    }
  },
  
  // 调试
  debug: process.env.NODE_ENV === 'development',
  
  // 后端选项
  backend: {
    loadPath: '/locales/{{lng}}/{{ns}}.json',
  },
  
  // 检测选项
  detection: {
    order: ['querystring', 'cookie', 'localStorage', 'navigator', 'htmlTag'],
    caches: ['localStorage', 'cookie'],
    excludeCacheFor: ['cimode'],
    cookieMinutes: 60 * 24 * 30, // 30 days
    cookieDomain: process.env.NODE_ENV === 'production' ? '.rwa-platform.com' : 'localhost',
  },
  
  // React选项
  react: {
    useSuspense: false,
    bindI18n: 'languageChanged',
    bindI18nStore: '',
    transEmptyNodeValue: '',
    transSupportBasicHtmlNodes: true,
    transKeepBasicHtmlNodesFor: ['br', 'strong', 'i'],
  },
};
