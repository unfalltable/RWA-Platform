import React, { useState, useEffect } from 'react';
import type { NextPage, GetStaticProps } from 'next';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { useTranslation } from 'next-i18next';
import { serverSideTranslations } from 'next-i18next/serverSideTranslations';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Eye, EyeOff, Mail, Lock, Wallet, ArrowRight } from 'lucide-react';

import { useAuth } from '@/contexts/AuthContext';
import { useWeb3 } from '@/contexts/Web3Context';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { SEO } from '@/components/common/SEO';
import { WalletConnectModal } from '@/components/web3/WalletConnectModal';
import { toast } from 'react-hot-toast';

// 表单验证schema
const loginSchema = z.object({
  email: z.string().email('请输入有效的邮箱地址'),
  password: z.string().min(6, '密码至少6位'),
  rememberMe: z.boolean().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

const LoginPage: NextPage = () => {
  const { t } = useTranslation('auth');
  const router = useRouter();
  const { login, isLoading, isAuthenticated } = useAuth();
  const { connectWallet, isConnected } = useWeb3();
  
  const [showPassword, setShowPassword] = useState(false);
  const [showWalletModal, setShowWalletModal] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  // 如果已登录，重定向到首页
  useEffect(() => {
    if (isAuthenticated) {
      const returnUrl = router.query.returnUrl as string || '/portfolio';
      router.push(returnUrl);
    }
  }, [isAuthenticated, router]);

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data);
    } catch (error) {
      // 错误已在AuthContext中处理
    }
  };

  const handleWeb3Login = async () => {
    try {
      setShowWalletModal(true);
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  return (
    <>
      <SEO
        title={t('login.title')}
        description={t('login.description')}
        noIndex
      />

      <div className="min-h-screen bg-gradient-to-br from-brand-50 via-white to-blue-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          {/* Logo和标题 */}
          <div className="text-center">
            <Link href="/" className="inline-flex items-center space-x-2">
              <div className="w-12 h-12 bg-brand-600 rounded-xl flex items-center justify-center">
                <span className="text-white font-bold text-lg">RWA</span>
              </div>
            </Link>
            <h2 className="mt-6 text-3xl font-bold text-gray-900">
              {t('login.welcome')}
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              {t('login.subtitle')}
            </p>
          </div>

          <Card className="p-8">
            {/* Web3登录选项 */}
            <div className="mb-6">
              <Button
                fullWidth
                variant="outline"
                size="lg"
                onClick={handleWeb3Login}
                leftIcon={<Wallet className="w-5 h-5" />}
                rightIcon={<ArrowRight className="w-5 h-5" />}
              >
                {t('login.web3Login')}
              </Button>
            </div>

            {/* 分割线 */}
            <div className="relative mb-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  {t('login.orContinueWith')}
                </span>
              </div>
            </div>

            {/* 邮箱登录表单 */}
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                  {t('login.email')}
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Mail className="h-5 w-5 text-gray-400" />
                  </div>
                  <Input
                    id="email"
                    type="email"
                    autoComplete="email"
                    className="pl-10"
                    placeholder={t('login.emailPlaceholder')}
                    {...register('email')}
                    error={errors.email?.message}
                  />
                </div>
              </div>

              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
                  {t('login.password')}
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-5 w-5 text-gray-400" />
                  </div>
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    autoComplete="current-password"
                    className="pl-10 pr-10"
                    placeholder={t('login.passwordPlaceholder')}
                    {...register('password')}
                    error={errors.password?.message}
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <EyeOff className="h-5 w-5 text-gray-400" />
                    ) : (
                      <Eye className="h-5 w-5 text-gray-400" />
                    )}
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <input
                    id="remember-me"
                    type="checkbox"
                    className="h-4 w-4 text-brand-600 focus:ring-brand-500 border-gray-300 rounded"
                    {...register('rememberMe')}
                  />
                  <label htmlFor="remember-me" className="ml-2 block text-sm text-gray-900">
                    {t('login.rememberMe')}
                  </label>
                </div>

                <div className="text-sm">
                  <Link
                    href="/forgot-password"
                    className="font-medium text-brand-600 hover:text-brand-500"
                  >
                    {t('login.forgotPassword')}
                  </Link>
                </div>
              </div>

              <Button
                type="submit"
                fullWidth
                size="lg"
                loading={isSubmitting || isLoading}
              >
                {t('login.signIn')}
              </Button>
            </form>

            {/* 注册链接 */}
            <div className="mt-6 text-center">
              <p className="text-sm text-gray-600">
                {t('login.noAccount')}{' '}
                <Link
                  href="/register"
                  className="font-medium text-brand-600 hover:text-brand-500"
                >
                  {t('login.signUp')}
                </Link>
              </p>
            </div>
          </Card>

          {/* 底部链接 */}
          <div className="text-center">
            <div className="flex justify-center space-x-6 text-sm text-gray-500">
              <Link href="/terms" className="hover:text-gray-700">
                {t('common.terms')}
              </Link>
              <Link href="/privacy" className="hover:text-gray-700">
                {t('common.privacy')}
              </Link>
              <Link href="/help" className="hover:text-gray-700">
                {t('common.help')}
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Web3钱包连接模态框 */}
      <WalletConnectModal
        isOpen={showWalletModal}
        onClose={() => setShowWalletModal(false)}
      />
    </>
  );
};

export const getStaticProps: GetStaticProps = async ({ locale }) => {
  return {
    props: {
      ...(await serverSideTranslations(locale ?? 'zh-CN', [
        'common',
        'auth',
      ])),
    },
  };
};

export default LoginPage;
