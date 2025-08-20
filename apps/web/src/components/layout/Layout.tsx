import React, { ReactNode } from 'react';
import { useRouter } from 'next/router';
import { Header } from './Header';
import { Footer } from './Footer';
import { Sidebar } from './Sidebar';
import { MobileMenu } from './MobileMenu';
import { useAuth } from '@/contexts/AuthContext';
import { cn } from '@/lib/utils';

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const router = useRouter();
  const { isAuthenticated } = useAuth();
  
  // 判断是否需要侧边栏的页面
  const needsSidebar = isAuthenticated && (
    router.pathname.startsWith('/portfolio') ||
    router.pathname.startsWith('/assets') ||
    router.pathname.startsWith('/channels') ||
    router.pathname.startsWith('/settings')
  );
  
  // 判断是否是全屏页面（如登录、注册）
  const isFullScreenPage = 
    router.pathname === '/login' ||
    router.pathname === '/register' ||
    router.pathname === '/forgot-password' ||
    router.pathname === '/reset-password';
  
  // 判断是否是着陆页
  const isLandingPage = router.pathname === '/';

  if (isFullScreenPage) {
    return (
      <div className="min-h-screen bg-gray-50">
        {children}
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 顶部导航 */}
      <Header />
      
      {/* 移动端菜单 */}
      <MobileMenu />
      
      {/* 主要内容区域 */}
      <div className={cn(
        "flex",
        needsSidebar && "pt-16" // 为固定头部留出空间
      )}>
        {/* 侧边栏 */}
        {needsSidebar && (
          <aside className="hidden lg:flex lg:flex-shrink-0">
            <div className="flex flex-col w-64">
              <Sidebar />
            </div>
          </aside>
        )}
        
        {/* 主内容 */}
        <main className={cn(
          "flex-1 min-w-0",
          needsSidebar ? "lg:pl-0" : "pt-16",
          isLandingPage && "pt-0"
        )}>
          <div className={cn(
            needsSidebar && "px-4 sm:px-6 lg:px-8 py-8",
            !needsSidebar && !isLandingPage && "container-responsive py-8"
          )}>
            {children}
          </div>
        </main>
      </div>
      
      {/* 底部 */}
      {!needsSidebar && <Footer />}
    </div>
  );
}
