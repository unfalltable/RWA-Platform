# RWA Platform - 稳定资产聚合与撮合平台

## 项目概述

面向全球华语用户的稳定资产（RWA）聚合与撮合平台，聚合资讯、行情、项目画像与风控评分；提供到第三方合规渠道的购买跳转与持仓/收益读取；不直接托管资产。

## 核心功能

- **资产聚合**: 美债/货币基金、稳定币、黄金/房产等 RWA 信息与指标
- **撮合服务**: 连接可购买渠道（CeFi、券商、合规 DeFi/发行方）
- **持仓分析**: 跨平台的持仓可视化、收益分析、风险预警
- **合规支持**: 税务/对账辅助、KYC/AML、地区化合规

## 技术架构

### 前端
- **Web**: Next.js 14 + TypeScript + Tailwind CSS + shadcn/ui
- **移动端**: Flutter (计划)
- **状态管理**: Zustand
- **国际化**: next-i18next

### 后端
- **API网关**: NestJS + GraphQL Federation
- **微服务**:
  - Go (数据采集、指标计算)
  - Java Spring Boot (合规、风控)
  - Python (数据科学、ML模型)
- **数据库**: PostgreSQL + Redis + ClickHouse
- **消息队列**: Kafka
- **实时计算**: Apache Flink

### 区块链
- **多链支持**: Ethereum, Arbitrum, Base, Solana, BSC, Tron
- **钱包集成**: Web3Auth + Account Abstraction (ERC-4337)
- **数据索引**: 自研索引器 + The Graph

### 基础设施
- **容器化**: Docker + Kubernetes
- **CI/CD**: GitHub Actions + ArgoCD
- **监控**: Prometheus + Grafana + OpenTelemetry
- **安全**: Vault + SAST/DAST

## 项目结构

```
rwa-platform/
├── apps/                    # 应用层
│   ├── web/                # Next.js Web应用
│   ├── mobile/             # Flutter移动应用
│   └── admin/              # 管理后台
├── services/               # 微服务
│   ├── api-gateway/        # API网关
│   ├── data-collector/     # 数据采集服务
│   ├── risk-engine/        # 风控引擎
│   ├── matching-service/   # 撮合服务
│   └── portfolio-service/  # 持仓服务
├── packages/               # 共享包
│   ├── types/              # TypeScript类型定义
│   ├── utils/              # 工具函数
│   ├── ui/                 # UI组件库
│   └── contracts/          # 智能合约
├── infrastructure/         # 基础设施
│   ├── k8s/               # Kubernetes配置
│   ├── terraform/         # Terraform配置
│   └── docker/            # Docker配置
└── docs/                  # 文档
```

## 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- Java 17+
- Python 3.11+
- Docker & Docker Compose
- Kubernetes (可选)

### 本地开发

1. 克隆项目
```bash
git clone <repository-url>
cd rwa-platform
```

2. 安装依赖
```bash
npm install
```

3. 启动开发环境
```bash
npm run dev
```

## 开发指南

详细的开发指南请参考 [docs/development.md](docs/development.md)

## 部署指南

详细的部署指南请参考 [docs/deployment.md](docs/deployment.md)

## 贡献指南

请参考 [CONTRIBUTING.md](CONTRIBUTING.md)

## 许可证

[MIT License](LICENSE)