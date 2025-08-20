# 开发指南

## 环境准备

### 系统要求
- Node.js 18+
- Go 1.21+
- Java 17+
- Python 3.11+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### 开发工具推荐
- VS Code + 推荐扩展
- Git
- Postman/Insomnia (API测试)
- DBeaver (数据库管理)

## 项目结构

```
rwa-platform/
├── apps/                    # 应用层
│   ├── web/                # Next.js Web应用
│   ├── mobile/             # Flutter移动应用
│   └── admin/              # 管理后台
├── services/               # 微服务
│   ├── api-gateway/        # API网关 (NestJS)
│   ├── data-collector/     # 数据采集服务 (Go)
│   ├── risk-engine/        # 风控引擎 (Java)
│   ├── matching-service/   # 撮合服务 (Go)
│   └── portfolio-service/  # 持仓服务 (Python)
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

### 1. 克隆项目
```bash
git clone <repository-url>
cd rwa-platform
```

### 2. 环境配置
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
vim .env
```

### 3. 安装依赖
```bash
npm install
```

### 4. 启动数据库
```bash
docker-compose up -d postgres redis clickhouse kafka
```

### 5. 数据库迁移
```bash
npm run db:migrate
```

### 6. 启动开发服务
```bash
# 启动所有服务
npm run dev

# 或者单独启动服务
npm run dev --workspace=@rwa-platform/web
npm run dev --workspace=@rwa-platform/api-gateway
```

## 开发规范

### 代码规范
- 使用 TypeScript
- 遵循 ESLint 规则
- 使用 Prettier 格式化
- 编写单元测试

### Git 规范
- 使用 Conventional Commits
- 分支命名: `feature/xxx`, `bugfix/xxx`, `hotfix/xxx`
- 提交前运行 `npm run lint` 和 `npm test`

### API 设计规范
- RESTful API 设计
- GraphQL Schema First
- 统一错误处理
- API 版本控制

## 测试

### 单元测试
```bash
npm run test
```

### 集成测试
```bash
npm run test:e2e
```

### 测试覆盖率
```bash
npm run test:coverage
```

## 部署

### 本地部署
```bash
docker-compose up
```

### 生产部署
```bash
# 构建镜像
npm run docker:build

# 部署到Kubernetes
npm run k8s:deploy
```

## 调试

### API调试
- 使用 Postman 集合
- GraphQL Playground: http://localhost:3000/graphql
- API文档: http://localhost:3000/api/docs

### 数据库调试
```bash
# 连接PostgreSQL
psql postgresql://postgres:postgres@localhost:5432/rwa_platform

# 连接Redis
redis-cli -h localhost -p 6379
```

### 日志查看
```bash
# 查看服务日志
docker-compose logs -f api-gateway

# 查看Kubernetes日志
kubectl logs -f deployment/api-gateway
```

## 常见问题

### Q: 端口冲突
A: 检查 `.env` 文件中的端口配置，确保没有被其他服务占用

### Q: 数据库连接失败
A: 确保数据库服务已启动，检查连接字符串是否正确

### Q: 依赖安装失败
A: 清除缓存后重新安装
```bash
rm -rf node_modules package-lock.json
npm install
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 相关链接

- [API文档](http://localhost:3000/api/docs)
- [GraphQL Playground](http://localhost:3000/graphql)
- [Storybook](http://localhost:6006)
- [监控面板](http://localhost:3001/monitoring)
