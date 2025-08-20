import React, { useState } from 'react';
import { Wallet, ChevronDown, Copy, ExternalLink, LogOut, Settings } from 'lucide-react';
import { useWeb3 } from '@/contexts/Web3Context';
import { Button } from '@/components/ui/Button';
import { Dropdown } from '@/components/ui/Dropdown';
import { Modal } from '@/components/ui/Modal';
import { WalletConnectModal } from './WalletConnectModal';
import { WalletDetailsModal } from './WalletDetailsModal';
import { truncateAddress, formatCurrency } from '@/lib/utils';
import { toast } from 'react-hot-toast';

export function WalletButton() {
  const {
    isConnected,
    currentWallet,
    wallets,
    connectWallet,
    disconnectWallet,
    setCurrentWallet,
  } = useWeb3();

  const [showConnectModal, setShowConnectModal] = useState(false);
  const [showDetailsModal, setShowDetailsModal] = useState(false);

  const handleConnect = async () => {
    setShowConnectModal(true);
  };

  const handleDisconnect = async () => {
    try {
      await disconnectWallet();
      toast.success('钱包已断开连接');
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  const handleCopyAddress = () => {
    if (currentWallet?.address) {
      navigator.clipboard.writeText(currentWallet.address);
      toast.success('地址已复制到剪贴板');
    }
  };

  const handleViewOnExplorer = () => {
    if (currentWallet?.address) {
      const explorerUrls: Record<string, string> = {
        ethereum: 'https://etherscan.io/address/',
        arbitrum: 'https://arbiscan.io/address/',
        base: 'https://basescan.org/address/',
        polygon: 'https://polygonscan.com/address/',
        bsc: 'https://bscscan.com/address/',
      };

      const baseUrl = explorerUrls[currentWallet.chain.toLowerCase()];
      if (baseUrl) {
        window.open(`${baseUrl}${currentWallet.address}`, '_blank');
      }
    }
  };

  if (!isConnected || !currentWallet) {
    return (
      <>
        <Button
          variant="outline"
          size="sm"
          onClick={handleConnect}
          leftIcon={<Wallet className="w-4 h-4" />}
        >
          连接钱包
        </Button>

        <WalletConnectModal
          isOpen={showConnectModal}
          onClose={() => setShowConnectModal(false)}
        />
      </>
    );
  }

  const walletMenuItems = [
    {
      label: '钱包详情',
      onClick: () => setShowDetailsModal(true),
      icon: Settings,
    },
    {
      label: '复制地址',
      onClick: handleCopyAddress,
      icon: Copy,
    },
    {
      label: '在浏览器中查看',
      onClick: handleViewOnExplorer,
      icon: ExternalLink,
    },
    ...(wallets.length > 1 ? [
      {
        type: 'divider' as const,
      },
      ...wallets
        .filter(wallet => wallet.address !== currentWallet.address)
        .map(wallet => ({
          label: `切换到 ${truncateAddress(wallet.address)}`,
          onClick: () => setCurrentWallet(wallet),
          icon: Wallet,
        })),
    ] : []),
    {
      type: 'divider' as const,
    },
    {
      label: '断开连接',
      onClick: handleDisconnect,
      icon: LogOut,
      className: 'text-red-600 hover:text-red-700',
    },
  ];

  return (
    <>
      <Dropdown
        trigger={
          <Button
            variant="outline"
            size="sm"
            className="flex items-center space-x-2 min-w-0"
          >
            <div className="flex items-center space-x-2 min-w-0">
              {/* 链图标 */}
              <div className="w-5 h-5 rounded-full bg-brand-100 flex items-center justify-center flex-shrink-0">
                <span className="text-xs font-medium text-brand-600">
                  {currentWallet.chain.charAt(0).toUpperCase()}
                </span>
              </div>
              
              {/* 钱包信息 */}
              <div className="flex flex-col items-start min-w-0">
                <span className="text-sm font-medium text-gray-900 truncate">
                  {truncateAddress(currentWallet.address)}
                </span>
                {currentWallet.balance && (
                  <span className="text-xs text-gray-500">
                    {parseFloat(currentWallet.balance).toFixed(4)} ETH
                  </span>
                )}
              </div>
            </div>
            
            <ChevronDown className="w-4 h-4 text-gray-400 flex-shrink-0" />
          </Button>
        }
        items={walletMenuItems}
        align="end"
      />

      <WalletDetailsModal
        isOpen={showDetailsModal}
        onClose={() => setShowDetailsModal(false)}
        wallet={currentWallet}
      />
    </>
  );
}
