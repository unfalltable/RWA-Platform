import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { useTranslation } from 'next-i18next';
import { 
  Menu, 
  X, 
  Search, 
  Bell, 
  User, 
  Settings, 
  LogOut,
  ChevronDown,
  Globe
} from 'lucide-react';

import { useAuth } from '@/contexts/AuthContext';
import { useWeb3 } from '@/contexts/Web3Context';
import { Button } from '@/components/ui/Button';
import { Avatar } from '@/components/ui/Avatar';
import { Dropdown } from '@/components/ui/Dropdown';
import { SearchBar } from '@/components/common/SearchBar';
import { NotificationDropdown } from '@/components/common/NotificationDropdown';
import { LanguageSelector } from '@/components/common/LanguageSelector';
import { ThemeToggle } from '@/components/common/ThemeToggle';
import { WalletButton } from '@/components/web3/WalletButton';
import { cn } from '@/lib/utils';

export function Header() {
  const { t } = useTranslation('navigation');
  const router = useRouter();
  const { isAuthenticated, user, logout } = useAuth();
  const { isConnected, currentWallet } = useWeb3();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const navigation = [
    { name: t('assets'), href: '/assets' },
    { name: t('channels'), href: '/channels' },
    { name: t('research'), href: '/research' },
    { name: t('about'), href: '/about' },
  ];

  const userMenuItems = [
    {
      label: t('profile'),
      href: '/settings/profile',
      icon: User,
    },
    {
      label: t('settings'),
      href: '/settings',
      icon: Settings,
    },
    {
      label: t('logout'),
      onClick: logout,
      icon: LogOut,
      className: 'text-red-600 hover:text-red-700',
    },
  ];

  return (
    <header className="bg-white shadow-sm border-b border-gray-200 fixed top-0 left-0 right-0 z-50">
      <div className="container-responsive">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-brand-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">RWA</span>
              </div>
              <span className="text-xl font-bold text-gray-900 hidden sm:block">
                RWA Platform
              </span>
            </Link>
          </div>

          {/* 桌面端导航 */}
          <nav className="hidden md:flex items-center space-x-8">
            {navigation.map((item) => (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  "text-sm font-medium transition-colors duration-200",
                  router.pathname.startsWith(item.href)
                    ? "text-brand-600"
                    : "text-gray-700 hover:text-brand-600"
                )}
              >
                {item.name}
              </Link>
            ))}
          </nav>

          {/* 搜索栏 */}
          <div className="hidden lg:flex flex-1 max-w-lg mx-8">
            <SearchBar placeholder={t('searchPlaceholder')} />
          </div>

          {/* 右侧操作区 */}
          <div className="flex items-center space-x-4">
            {/* 搜索按钮（移动端） */}
            <Button
              variant="ghost"
              size="sm"
              className="lg:hidden"
              onClick={() => {
                // 打开搜索模态框
              }}
            >
              <Search className="w-5 h-5" />
            </Button>

            {/* 语言选择器 */}
            <LanguageSelector />

            {/* 主题切换 */}
            <ThemeToggle />

            {isAuthenticated ? (
              <>
                {/* 通知 */}
                <NotificationDropdown />

                {/* Web3钱包 */}
                <WalletButton />

                {/* 用户菜单 */}
                <Dropdown
                  trigger={
                    <Button variant="ghost" size="sm" className="flex items-center space-x-2">
                      <Avatar
                        src={user?.profile?.avatar}
                        alt={user?.profile?.displayName || user?.email}
                        size="sm"
                      />
                      <span className="hidden sm:block text-sm font-medium">
                        {user?.profile?.displayName || user?.email}
                      </span>
                      <ChevronDown className="w-4 h-4" />
                    </Button>
                  }
                  items={userMenuItems}
                />
              </>
            ) : (
              <>
                {/* 未登录状态 */}
                <Link href="/login">
                  <Button variant="ghost" size="sm">
                    {t('login')}
                  </Button>
                </Link>
                <Link href="/register">
                  <Button size="sm">
                    {t('register')}
                  </Button>
                </Link>
              </>
            )}

            {/* 移动端菜单按钮 */}
            <Button
              variant="ghost"
              size="sm"
              className="md:hidden"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            >
              {isMobileMenuOpen ? (
                <X className="w-5 h-5" />
              ) : (
                <Menu className="w-5 h-5" />
              )}
            </Button>
          </div>
        </div>
      </div>

      {/* 移动端菜单 */}
      {isMobileMenuOpen && (
        <div className="md:hidden bg-white border-t border-gray-200">
          <div className="px-4 py-2 space-y-1">
            {/* 搜索栏 */}
            <div className="py-2">
              <SearchBar placeholder={t('searchPlaceholder')} />
            </div>

            {/* 导航链接 */}
            {navigation.map((item) => (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  "block px-3 py-2 text-base font-medium rounded-md transition-colors duration-200",
                  router.pathname.startsWith(item.href)
                    ? "text-brand-600 bg-brand-50"
                    : "text-gray-700 hover:text-brand-600 hover:bg-gray-50"
                )}
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {item.name}
              </Link>
            ))}

            {/* 用户相关链接 */}
            {isAuthenticated && (
              <div className="pt-4 border-t border-gray-200">
                <Link
                  href="/portfolio"
                  className="block px-3 py-2 text-base font-medium text-gray-700 hover:text-brand-600 hover:bg-gray-50 rounded-md"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {t('portfolio')}
                </Link>
                <Link
                  href="/settings"
                  className="block px-3 py-2 text-base font-medium text-gray-700 hover:text-brand-600 hover:bg-gray-50 rounded-md"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {t('settings')}
                </Link>
              </div>
            )}
          </div>
        </div>
      )}
    </header>
  );
}
