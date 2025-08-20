import { Web3Auth } from '@web3auth/modal';
import { WALLET_ADAPTERS, CHAIN_NAMESPACES, IProvider } from '@web3auth/base';
import { EthereumProvider } from '@web3auth/ethereum-provider';
import { web3AuthConfig, getAdapters, chainConfigs, chainSwitchConfig } from './config';
import { authApi } from '@/lib/api/auth';
import { toast } from 'react-hot-toast';

export interface Web3AuthUser {
  email?: string;
  name?: string;
  profileImage?: string;
  aggregateVerifier?: string;
  verifier?: string;
  verifierId?: string;
  typeOfLogin?: string;
}

export interface WalletInfo {
  address: string;
  balance: string;
  chainId: number;
  chainName: string;
}

class Web3AuthService {
  private web3auth: Web3Auth | null = null;
  private provider: IProvider | null = null;
  private isInitialized = false;

  /**
   * 初始化Web3Auth
   */
  async init(): Promise<void> {
    try {
      if (this.isInitialized) return;

      this.web3auth = new Web3Auth(web3AuthConfig);

      // 配置适配器
      const adapters = getAdapters();
      adapters.forEach(adapter => {
        this.web3auth!.configureAdapter(adapter);
      });

      await this.web3auth.initModal();
      this.provider = this.web3auth.provider;
      this.isInitialized = true;

      console.log('Web3Auth initialized successfully');
    } catch (error) {
      console.error('Web3Auth initialization failed:', error);
      throw new Error('Web3Auth初始化失败');
    }
  }

  /**
   * 检查是否已连接
   */
  isConnected(): boolean {
    return this.web3auth?.connected || false;
  }

  /**
   * 获取当前用户信息
   */
  async getUserInfo(): Promise<Web3AuthUser | null> {
    if (!this.web3auth?.connected) return null;

    try {
      const userInfo = await this.web3auth.getUserInfo();
      return userInfo as Web3AuthUser;
    } catch (error) {
      console.error('Failed to get user info:', error);
      return null;
    }
  }

  /**
   * 连接钱包
   */
  async connect(loginProvider?: string): Promise<{
    user: Web3AuthUser;
    wallet: WalletInfo;
  }> {
    try {
      if (!this.isInitialized) {
        await this.init();
      }

      if (!this.web3auth) {
        throw new Error('Web3Auth not initialized');
      }

      // 连接钱包
      this.provider = await this.web3auth.connect();
      
      if (!this.provider) {
        throw new Error('Failed to connect wallet');
      }

      // 获取用户信息
      const userInfo = await this.getUserInfo();
      if (!userInfo) {
        throw new Error('Failed to get user info');
      }

      // 获取钱包信息
      const walletInfo = await this.getWalletInfo();

      return {
        user: userInfo,
        wallet: walletInfo,
      };
    } catch (error: any) {
      console.error('Wallet connection failed:', error);
      throw new Error(error.message || '钱包连接失败');
    }
  }

  /**
   * 断开连接
   */
  async disconnect(): Promise<void> {
    try {
      if (this.web3auth?.connected) {
        await this.web3auth.logout();
      }
      this.provider = null;
    } catch (error) {
      console.error('Disconnect failed:', error);
      throw new Error('断开连接失败');
    }
  }

  /**
   * 获取钱包信息
   */
  async getWalletInfo(): Promise<WalletInfo> {
    if (!this.provider) {
      throw new Error('Wallet not connected');
    }

    try {
      const ethProvider = new EthereumProvider({ provider: this.provider });
      
      // 获取账户地址
      const accounts = await ethProvider.request({ method: 'eth_accounts' });
      const address = accounts[0];

      // 获取链ID
      const chainId = await ethProvider.request({ method: 'eth_chainId' });
      const chainIdNumber = parseInt(chainId, 16);

      // 获取余额
      const balance = await ethProvider.request({
        method: 'eth_getBalance',
        params: [address, 'latest'],
      });
      const balanceInEth = parseInt(balance, 16) / Math.pow(10, 18);

      // 获取链名称
      const chainName = this.getChainName(chainIdNumber);

      return {
        address,
        balance: balanceInEth.toFixed(4),
        chainId: chainIdNumber,
        chainName,
      };
    } catch (error) {
      console.error('Failed to get wallet info:', error);
      throw new Error('获取钱包信息失败');
    }
  }

  /**
   * 切换网络
   */
  async switchChain(chainKey: string): Promise<void> {
    if (!this.provider) {
      throw new Error('Wallet not connected');
    }

    try {
      const ethProvider = new EthereumProvider({ provider: this.provider });
      const chainConfig = chainSwitchConfig[chainKey];

      if (!chainConfig) {
        throw new Error('Unsupported chain');
      }

      try {
        // 尝试切换到目标网络
        await ethProvider.request({
          method: 'wallet_switchEthereumChain',
          params: [{ chainId: chainConfig.chainId }],
        });
      } catch (switchError: any) {
        // 如果网络不存在，尝试添加网络
        if (switchError.code === 4902) {
          await ethProvider.request({
            method: 'wallet_addEthereumChain',
            params: [chainConfig],
          });
        } else {
          throw switchError;
        }
      }

      toast.success(`已切换到 ${chainConfig.chainName}`);
    } catch (error: any) {
      console.error('Chain switch failed:', error);
      throw new Error(`网络切换失败: ${error.message}`);
    }
  }

  /**
   * 签名消息
   */
  async signMessage(message: string): Promise<string> {
    if (!this.provider) {
      throw new Error('Wallet not connected');
    }

    try {
      const ethProvider = new EthereumProvider({ provider: this.provider });
      const accounts = await ethProvider.request({ method: 'eth_accounts' });
      
      const signature = await ethProvider.request({
        method: 'personal_sign',
        params: [message, accounts[0]],
      });

      return signature;
    } catch (error: any) {
      console.error('Message signing failed:', error);
      throw new Error(`签名失败: ${error.message}`);
    }
  }

  /**
   * 发送交易
   */
  async sendTransaction(params: {
    to: string;
    value?: string;
    data?: string;
    gasLimit?: string;
    gasPrice?: string;
  }): Promise<string> {
    if (!this.provider) {
      throw new Error('Wallet not connected');
    }

    try {
      const ethProvider = new EthereumProvider({ provider: this.provider });
      const accounts = await ethProvider.request({ method: 'eth_accounts' });

      const txParams = {
        from: accounts[0],
        to: params.to,
        value: params.value || '0x0',
        data: params.data || '0x',
        gas: params.gasLimit,
        gasPrice: params.gasPrice,
      };

      const txHash = await ethProvider.request({
        method: 'eth_sendTransaction',
        params: [txParams],
      });

      return txHash;
    } catch (error: any) {
      console.error('Transaction failed:', error);
      throw new Error(`交易失败: ${error.message}`);
    }
  }

  /**
   * 获取私钥（仅用于特殊场景）
   */
  async getPrivateKey(): Promise<string> {
    if (!this.provider) {
      throw new Error('Wallet not connected');
    }

    try {
      const privateKey = await this.provider.request({
        method: 'eth_private_key',
      });

      return privateKey;
    } catch (error: any) {
      console.error('Failed to get private key:', error);
      throw new Error('获取私钥失败');
    }
  }

  /**
   * Web3登录到后端
   */
  async loginToBackend(): Promise<void> {
    try {
      const walletInfo = await this.getWalletInfo();
      
      // 获取登录消息
      const { message, nonce } = await authApi.getWeb3LoginMessage(
        walletInfo.address,
        walletInfo.chainId
      );

      // 签名消息
      const signature = await this.signMessage(message);

      // 发送登录请求
      await authApi.web3Login({
        address: walletInfo.address,
        signature,
        message,
        chainId: walletInfo.chainId,
      });

      toast.success('Web3登录成功');
    } catch (error: any) {
      console.error('Web3 login failed:', error);
      throw new Error(`Web3登录失败: ${error.message}`);
    }
  }

  /**
   * 获取链名称
   */
  private getChainName(chainId: number): string {
    const chainMap: Record<number, string> = {
      1: 'Ethereum',
      42161: 'Arbitrum',
      8453: 'Base',
      137: 'Polygon',
      56: 'BSC',
    };

    return chainMap[chainId] || `Chain ${chainId}`;
  }

  /**
   * 获取提供商实例
   */
  getProvider(): IProvider | null {
    return this.provider;
  }

  /**
   * 获取Web3Auth实例
   */
  getWeb3Auth(): Web3Auth | null {
    return this.web3auth;
  }
}

// 导出单例实例
export const web3AuthService = new Web3AuthService();
