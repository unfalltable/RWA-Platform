import React, { useState } from 'react';
import { X, Wallet, Mail, Smartphone, Globe, Shield, ArrowRight } from 'lucide-react';
import { useWeb3 } from '@/contexts/Web3Context';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { supportedWallets, loginProviders } from '@/lib/web3auth/config';
import { toast } from 'react-hot-toast';

interface WalletConnectModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function WalletConnectModal({ isOpen, onClose }: WalletConnectModalProps) {
  const { connectWallet, isConnecting } = useWeb3();
  const [selectedMethod, setSelectedMethod] = useState<'wallet' | 'social'>('wallet');

  const handleWalletConnect = async (walletId: string) => {
    try {
      await connectWallet(walletId);
      onClose();
      toast.success('钱包连接成功');
    } catch (error: any) {
      toast.error(error.message || '钱包连接失败');
    }
  };

  const handleSocialConnect = async (provider: string) => {
    try {
      await connectWallet(provider);
      onClose();
      toast.success('登录成功');
    } catch (error: any) {
      toast.error(error.message || '登录失败');
    }
  };

  const socialProviders = [
    {
      id: 'google',
      name: 'Google',
      icon: <Globe className="w-5 h-5" />,
      description: '使用Google账户登录',
      color: 'bg-red-50 text-red-600 border-red-200',
    },
    {
      id: 'apple',
      name: 'Apple',
      icon: <Smartphone className="w-5 h-5" />,
      description: '使用Apple ID登录',
      color: 'bg-gray-50 text-gray-600 border-gray-200',
    },
    {
      id: 'email',
      name: '邮箱',
      icon: <Mail className="w-5 h-5" />,
      description: '使用邮箱无密码登录',
      color: 'bg-blue-50 text-blue-600 border-blue-200',
    },
  ];

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <div className="p-6">
        {/* 头部 */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">
            连接钱包
          </h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* 连接方式选择 */}
        <div className="flex space-x-1 mb-6 bg-gray-100 rounded-lg p-1">
          <button
            className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
              selectedMethod === 'wallet'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
            onClick={() => setSelectedMethod('wallet')}
          >
            <Wallet className="w-4 h-4 inline mr-2" />
            钱包连接
          </button>
          <button
            className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
              selectedMethod === 'social'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-600 hover:text-gray-900'
            }`}
            onClick={() => setSelectedMethod('social')}
          >
            <Shield className="w-4 h-4 inline mr-2" />
            社交登录
          </button>
        </div>

        {/* 钱包连接选项 */}
        {selectedMethod === 'wallet' && (
          <div className="space-y-3">
            <p className="text-sm text-gray-600 mb-4">
              选择您偏好的钱包来连接到RWA Platform
            </p>
            
            {supportedWallets.map((wallet) => (
              <Card
                key={wallet.id}
                className="p-4 hover:shadow-md transition-all duration-200 cursor-pointer border-2 hover:border-brand-200"
                onClick={() => handleWalletConnect(wallet.id)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center">
                      <img
                        src={wallet.icon}
                        alt={wallet.name}
                        className="w-6 h-6"
                        onError={(e) => {
                          (e.target as HTMLImageElement).src = '/wallets/default.svg';
                        }}
                      />
                    </div>
                    <div>
                      <h3 className="font-medium text-gray-900">{wallet.name}</h3>
                      <p className="text-sm text-gray-600">{wallet.description}</p>
                    </div>
                  </div>
                  <ArrowRight className="w-5 h-5 text-gray-400" />
                </div>
              </Card>
            ))}

            <div className="mt-6 p-4 bg-blue-50 rounded-lg">
              <div className="flex items-start space-x-3">
                <Shield className="w-5 h-5 text-blue-600 mt-0.5" />
                <div>
                  <h4 className="text-sm font-medium text-blue-900">安全提示</h4>
                  <p className="text-sm text-blue-700 mt-1">
                    我们不会存储您的私钥。所有交易都需要您的确认。
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 社交登录选项 */}
        {selectedMethod === 'social' && (
          <div className="space-y-3">
            <p className="text-sm text-gray-600 mb-4">
              使用社交账户快速登录，我们会为您自动创建钱包
            </p>
            
            {socialProviders.map((provider) => (
              <Card
                key={provider.id}
                className="p-4 hover:shadow-md transition-all duration-200 cursor-pointer border-2 hover:border-brand-200"
                onClick={() => handleSocialConnect(provider.id)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${provider.color}`}>
                      {provider.icon}
                    </div>
                    <div>
                      <h3 className="font-medium text-gray-900">{provider.name}</h3>
                      <p className="text-sm text-gray-600">{provider.description}</p>
                    </div>
                  </div>
                  <ArrowRight className="w-5 h-5 text-gray-400" />
                </div>
              </Card>
            ))}

            <div className="mt-6 p-4 bg-green-50 rounded-lg">
              <div className="flex items-start space-x-3">
                <Shield className="w-5 h-5 text-green-600 mt-0.5" />
                <div>
                  <h4 className="text-sm font-medium text-green-900">Web3Auth技术</h4>
                  <p className="text-sm text-green-700 mt-1">
                    基于Web3Auth技术，为您提供安全便捷的Web3体验，无需记住助记词。
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 加载状态 */}
        {isConnecting && (
          <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center rounded-lg">
            <div className="text-center">
              <div className="loading-spinner mx-auto mb-2" />
              <p className="text-sm text-gray-600">正在连接...</p>
            </div>
          </div>
        )}

        {/* 底部说明 */}
        <div className="mt-6 pt-4 border-t border-gray-200">
          <p className="text-xs text-gray-500 text-center">
            连接钱包即表示您同意我们的
            <a href="/terms" className="text-brand-600 hover:underline ml-1">服务条款</a>
            和
            <a href="/privacy" className="text-brand-600 hover:underline ml-1">隐私政策</a>
          </p>
        </div>
      </div>
    </Modal>
  );
}
