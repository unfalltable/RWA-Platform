import { Web3AuthOptions } from '@web3auth/modal';
import { CHAIN_NAMESPACES, CustomChainConfig } from '@web3auth/base';
import { EthereumPrivateKeyProvider } from '@web3auth/ethereum-provider';
import { MetamaskAdapter } from '@web3auth/metamask-adapter';
import { WalletConnectV2Adapter } from '@web3auth/wallet-connect-v2-adapter';
import { TorusWalletAdapter } from '@web3auth/torus-evm-adapter';

// 支持的链配置
export const chainConfigs: Record<string, CustomChainConfig> = {
  ethereum: {
    chainNamespace: CHAIN_NAMESPACES.EIP155,
    chainId: '0x1',
    rpcTarget: process.env.NEXT_PUBLIC_ETHEREUM_RPC_URL || 'https://rpc.ankr.com/eth',
    displayName: 'Ethereum Mainnet',
    blockExplorer: 'https://etherscan.io',
    ticker: 'ETH',
    tickerName: 'Ethereum',
  },
  arbitrum: {
    chainNamespace: CHAIN_NAMESPACES.EIP155,
    chainId: '0xa4b1',
    rpcTarget: process.env.NEXT_PUBLIC_ARBITRUM_RPC_URL || 'https://rpc.ankr.com/arbitrum',
    displayName: 'Arbitrum One',
    blockExplorer: 'https://arbiscan.io',
    ticker: 'ETH',
    tickerName: 'Ethereum',
  },
  base: {
    chainNamespace: CHAIN_NAMESPACES.EIP155,
    chainId: '0x2105',
    rpcTarget: process.env.NEXT_PUBLIC_BASE_RPC_URL || 'https://mainnet.base.org',
    displayName: 'Base',
    blockExplorer: 'https://basescan.org',
    ticker: 'ETH',
    tickerName: 'Ethereum',
  },
  polygon: {
    chainNamespace: CHAIN_NAMESPACES.EIP155,
    chainId: '0x89',
    rpcTarget: process.env.NEXT_PUBLIC_POLYGON_RPC_URL || 'https://rpc.ankr.com/polygon',
    displayName: 'Polygon',
    blockExplorer: 'https://polygonscan.com',
    ticker: 'MATIC',
    tickerName: 'Polygon',
  },
  bsc: {
    chainNamespace: CHAIN_NAMESPACES.EIP155,
    chainId: '0x38',
    rpcTarget: process.env.NEXT_PUBLIC_BSC_RPC_URL || 'https://rpc.ankr.com/bsc',
    displayName: 'BNB Smart Chain',
    blockExplorer: 'https://bscscan.com',
    ticker: 'BNB',
    tickerName: 'BNB',
  },
};

// Web3Auth配置
export const web3AuthConfig: Web3AuthOptions = {
  clientId: process.env.NEXT_PUBLIC_WEB3AUTH_CLIENT_ID || '',
  web3AuthNetwork: process.env.NODE_ENV === 'production' ? 'mainnet' : 'testnet',
  chainConfig: chainConfigs.ethereum,
  uiConfig: {
    appName: 'RWA Platform',
    appUrl: process.env.NEXT_PUBLIC_APP_URL || 'https://rwa-platform.com',
    theme: {
      primary: '#0ea5e9',
    },
    mode: 'light',
    logoLight: '/logo-light.png',
    logoDark: '/logo-dark.png',
    defaultLanguage: 'zh',
    loginMethodsOrder: ['google', 'apple', 'twitter', 'discord', 'email_passwordless'],
  },
  privateKeyProvider: new EthereumPrivateKeyProvider({
    config: {
      chainConfig: chainConfigs.ethereum,
    },
  }),
  sessionTime: 86400, // 24小时
  storageKey: 'local',
};

// 适配器配置
export const getAdapters = () => {
  const adapters = [];

  // MetaMask适配器
  const metamaskAdapter = new MetamaskAdapter({
    clientId: web3AuthConfig.clientId,
    sessionTime: 86400,
    web3AuthNetwork: web3AuthConfig.web3AuthNetwork,
    chainConfig: chainConfigs.ethereum,
  });
  adapters.push(metamaskAdapter);

  // WalletConnect适配器
  const walletConnectV2Adapter = new WalletConnectV2Adapter({
    adapterSettings: {
      qrcodeModal: null,
      walletConnectInitOptions: {
        projectId: process.env.NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID || '',
        metadata: {
          name: 'RWA Platform',
          description: '稳定资产聚合与撮合平台',
          url: process.env.NEXT_PUBLIC_APP_URL || 'https://rwa-platform.com',
          icons: ['/logo.png'],
        },
      },
    },
    loginSettings: {
      mfaLevel: 'default',
    },
    clientId: web3AuthConfig.clientId,
    sessionTime: 86400,
    web3AuthNetwork: web3AuthConfig.web3AuthNetwork,
    chainConfig: chainConfigs.ethereum,
  });
  adapters.push(walletConnectV2Adapter);

  // Torus钱包适配器
  const torusWalletAdapter = new TorusWalletAdapter({
    adapterSettings: {
      buttonPosition: 'bottom-left',
    },
    loginSettings: {
      verifier: 'rwa-platform',
    },
    initParams: {
      buildEnv: process.env.NODE_ENV === 'production' ? 'production' : 'testing',
    },
    clientId: web3AuthConfig.clientId,
    sessionTime: 86400,
    web3AuthNetwork: web3AuthConfig.web3AuthNetwork,
    chainConfig: chainConfigs.ethereum,
  });
  adapters.push(torusWalletAdapter);

  return adapters;
};

// 登录提供商配置
export const loginProviders = {
  google: {
    name: 'Google',
    verifier: 'rwa-google',
    typeOfLogin: 'google',
    clientId: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || '',
  },
  apple: {
    name: 'Apple',
    verifier: 'rwa-apple',
    typeOfLogin: 'apple',
    clientId: process.env.NEXT_PUBLIC_APPLE_CLIENT_ID || '',
  },
  twitter: {
    name: 'Twitter',
    verifier: 'rwa-twitter',
    typeOfLogin: 'twitter',
    clientId: process.env.NEXT_PUBLIC_TWITTER_CLIENT_ID || '',
  },
  discord: {
    name: 'Discord',
    verifier: 'rwa-discord',
    typeOfLogin: 'discord',
    clientId: process.env.NEXT_PUBLIC_DISCORD_CLIENT_ID || '',
  },
  email: {
    name: 'Email',
    verifier: 'rwa-email',
    typeOfLogin: 'email_passwordless',
  },
};

// 支持的钱包列表
export const supportedWallets = [
  {
    id: 'metamask',
    name: 'MetaMask',
    icon: '/wallets/metamask.svg',
    description: '最受欢迎的以太坊钱包',
    downloadUrl: 'https://metamask.io/download/',
  },
  {
    id: 'walletconnect',
    name: 'WalletConnect',
    icon: '/wallets/walletconnect.svg',
    description: '连接多种移动钱包',
    downloadUrl: 'https://walletconnect.com/',
  },
  {
    id: 'coinbase',
    name: 'Coinbase Wallet',
    icon: '/wallets/coinbase.svg',
    description: 'Coinbase官方钱包',
    downloadUrl: 'https://wallet.coinbase.com/',
  },
  {
    id: 'trust',
    name: 'Trust Wallet',
    icon: '/wallets/trust.svg',
    description: '安全的多链钱包',
    downloadUrl: 'https://trustwallet.com/',
  },
];

// 链切换配置
export const chainSwitchConfig = {
  ethereum: {
    chainId: '0x1',
    chainName: 'Ethereum Mainnet',
    nativeCurrency: {
      name: 'Ethereum',
      symbol: 'ETH',
      decimals: 18,
    },
    rpcUrls: ['https://rpc.ankr.com/eth'],
    blockExplorerUrls: ['https://etherscan.io'],
  },
  arbitrum: {
    chainId: '0xa4b1',
    chainName: 'Arbitrum One',
    nativeCurrency: {
      name: 'Ethereum',
      symbol: 'ETH',
      decimals: 18,
    },
    rpcUrls: ['https://rpc.ankr.com/arbitrum'],
    blockExplorerUrls: ['https://arbiscan.io'],
  },
  base: {
    chainId: '0x2105',
    chainName: 'Base',
    nativeCurrency: {
      name: 'Ethereum',
      symbol: 'ETH',
      decimals: 18,
    },
    rpcUrls: ['https://mainnet.base.org'],
    blockExplorerUrls: ['https://basescan.org'],
  },
  polygon: {
    chainId: '0x89',
    chainName: 'Polygon',
    nativeCurrency: {
      name: 'MATIC',
      symbol: 'MATIC',
      decimals: 18,
    },
    rpcUrls: ['https://rpc.ankr.com/polygon'],
    blockExplorerUrls: ['https://polygonscan.com'],
  },
  bsc: {
    chainId: '0x38',
    chainName: 'BNB Smart Chain',
    nativeCurrency: {
      name: 'BNB',
      symbol: 'BNB',
      decimals: 18,
    },
    rpcUrls: ['https://rpc.ankr.com/bsc'],
    blockExplorerUrls: ['https://bscscan.com'],
  },
};
