// 通用类型定义

export type UUID = string;
export type Timestamp = number;
export type Address = string;
export type ChainId = number;

// 支持的区块链网络
export enum Chain {
  ETHEREUM = 'ethereum',
  ARBITRUM = 'arbitrum',
  BASE = 'base',
  SOLANA = 'solana',
  BSC = 'bsc',
  TRON = 'tron',
  POLYGON = 'polygon'
}

// 支持的货币
export enum Currency {
  USD = 'USD',
  EUR = 'EUR',
  CNY = 'CNY',
  JPY = 'JPY',
  GBP = 'GBP'
}

// 地区代码
export enum Region {
  US = 'US',
  EU = 'EU',
  CN = 'CN',
  SG = 'SG',
  HK = 'HK',
  JP = 'JP'
}

// 分页参数
export interface PaginationParams {
  page: number;
  limit: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// 分页响应
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

// API响应包装
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  timestamp: Timestamp;
}

// 价格信息
export interface Price {
  value: number;
  currency: Currency;
  timestamp: Timestamp;
  source?: string;
}

// 百分比变化
export interface PercentageChange {
  value: number;
  period: '1h' | '24h' | '7d' | '30d' | '1y';
}

// 地址信息
export interface AddressInfo {
  address: Address;
  chain: Chain;
  label?: string;
  isContract?: boolean;
}
