# 筑价｜建材比价平台

`全栈Web应用` · 面向装修业主和施工队的建材价格对比平台。聚合多商家报价、供货状态和价格记录，让采购决策有清晰依据。

## Docker Compose 一键启动（推荐）

> 首次启动前，请先复制环境变量文件并按需修改密码与端口。

```bash
cd "农业与生活服务主题项目提示词/ld-325"
cp .env.example .env
docker compose up -d
```

检查服务状态：

```bash
docker compose ps
curl http://localhost:19625/healthz
```

停止并移除服务：

```bash
docker compose down
```

## 访问地址

| 服务 | 地址 |
| --- | --- |
| Web 页面 | http://localhost:18625 |
| 后端健康检查 | http://localhost:19625/healthz |
| 商品 API | http://localhost:19625/api/v1/products |
| OpenAPI 定义 | [`backend/api/openapi.yaml`](backend/api/openapi.yaml) |

## 主要功能

- **分类与检索**：覆盖瓷砖、地板、涂料、卫浴、五金、门窗、灯具、管材；支持关键词、分类和排序接口。
- **多商家报价**：同款材料显示店铺、单价、起订量、运费、交货期及库存状态，最低价高亮。
- **收藏与对比**：收藏进入“本周采购”文件夹；可同时把 2–4 款材料纳入对比清单。
- **价格趋势**：读取报价历史，展示 30/90 天或 1 年区间的最高、最低和平均价；前端以 ECharts 绘制 30 天图表。
- **价格预警**：每款材料仅保留一条订阅，支持目标价与降幅百分比。商家改价经管理员审核通过后，符合条件的订阅触发一次并记录触发价与时间；提醒页展示当前最低价、触发价和状态。
- **报价修改审核**：供应商提交单价、运费、货期与库存状态变更；同一报价有待审修改时新提交不入库（409）。审核基于报价版本，版本已变化则将本次修改标记为失效并返回冲突；通过后旧价写入价格历史并触发价格提醒，供应商可在工作台查看审核结果。
- **供应商管理**：供应商资质和审核状态可查询；管理员审核、供应商库存状态更新接口已保留。
- **装修预算**：按客厅、厨房、卫生间和面积基于市场均价试算，结果可保存，前端提供导出入口。

## 本地开发（备选）

### 1. 基础依赖

启动 PostgreSQL 和 Redis（或直接使用 Docker Compose 的 `db`、`redis` 服务）：

```bash
docker compose up -d db redis
```

### 2. 后端

```bash
cd backend
go mod tidy
go run ./cmd/server
```

后端监听 `http://localhost:19625`（本地默认值来自配置；若未设置 `SERVER_PORT` 则为 `8080`）。为保持与根目录 `.env` 的端口一致，可执行：

```bash
SERVER_PORT=19625 DB_HOST=localhost REDIS_HOST=localhost go run ./cmd/server
```

### 3. 前端

```bash
cd frontend
npm ci
npm run dev -- -p 18625
```

本地开发的 Next.js 页面会请求同源 `/api/v1`。推荐通过 Docker 前端 Nginx 运行以获得 API 反向代理；如单独开发前端，请在本地增加等效反向代理。

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 前端 | Next.js 14、TypeScript、Tailwind CSS、shadcn/ui 风格基础组件、ECharts |
| 后端 | **Go 1.22 + Gin + GORM** |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7（go-redis/v9） |
| 认证 | JWT（golang-jwt/jwt/v5；支持 Bearer Token 解析与 admin/supplier/user RBAC；未携带令牌时使用 demo-user 浏览演示数据） |
| 部署 | Docker Compose、Nginx、多阶段 Dockerfile |

## API 概览

所有业务响应统一为：`{"code":0,"message":"ok","data":...}`，业务路由统一前缀为 `/api/v1`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/auth/demo-token` | 签发演示身份令牌（admin/supplier/user） |
| GET | `/api/v1/products?q=&category=&sort=&page=&page_size=` | 搜索建材；`sort` 支持 `price`、`sales`、`rating` |
| GET | `/api/v1/products/:id` | 建材详情及报价 |
| POST | `/api/v1/products/compare` | 批量比较，body：`{"ids":[1,2]}` |
| GET | `/api/v1/products/:id/offers` | 某建材商家报价 |
| GET | `/api/v1/products/:id/trend?range=30d` | 价格趋势，支持 `30d`、`90d`、`1y` |
| GET/POST | `/api/v1/favorites` | 收藏列表 / 添加收藏 |
| GET/POST | `/api/v1/alerts` | 价格提醒列表（最低价/触发价/状态） / 创建订阅（同款仅一条） |
| POST | `/api/v1/budgets` | 保存预算试算 |
| GET | `/api/v1/suppliers` | 查询供应商 |
| GET | `/api/v1/supplier/offers` | 供应商查看名下报价（supplier/admin） |
| POST | `/api/v1/supplier/offers/:id/changes` | 提交报价修改；有待审修改时返回 409 |
| GET | `/api/v1/supplier/offer-changes` | 供应商查看修改与审核结果 |
| GET | `/api/v1/admin/offer-changes?status=pending` | 管理员审核队列（admin） |
| POST | `/api/v1/admin/offer-changes/:id/review` | 审核报价修改；版本冲突返回 409 并标记失效 |
| PATCH | `/api/v1/admin/suppliers/:id/status` | 审核供应商（admin 角色） |
| PATCH | `/api/v1/supplier/offers/:id/status` | 更新报价库存状态（supplier/admin 角色） |

### 报价修改与价格提醒流程

1. 供应商在「供应商工作台」（`/supplier`）对名下报价提交新单价、运费、货期和库存状态。
2. 同一报价已有待审核修改时，后续提交不入库并返回 `409`；审核结束后才能再次提交。
3. 管理员在「审核台」（`/admin`）基于报价版本审核。审核时报价版本已变化，则该修改被标记为 `stale`（失效）并返回 `409` 冲突。
4. 审核通过：旧单价写入价格历史、报价版本 `+1`，随后以该建材当前最低价评估订阅——达到目标价或订阅时设定的降幅时，订阅触发一次，记录触发价与时间。
5. 用户在「价格提醒」页（`/alerts`）查看当前最低价、触发价、监控/已触发状态和触发时间。

> 演示部署没有账号密码体系，供应商与管理员页面通过右上角入口调用 `/api/v1/auth/demo-token` 领取演示令牌（供应商示例身份为种子中的「筑家优选旗舰店」，subject `1`）。

## 目录结构

```text
ld-325/
├── docker-compose.yml              # 前端、后端、PostgreSQL、Redis 编排
├── .env.example                    # 可复制的部署变量样例
├── frontend/
│   ├── app/                        # Next.js 页面与全局样式
│   ├── components/                 # 目录、对比、趋势、预算等细分组件
│   ├── lib/                        # API 客户端、类型、格式化工具
│   ├── Dockerfile
│   └── nginx.conf                  # SPA 路由与 /api/ 反向代理
└── backend/
    ├── cmd/server/                 # 仅负责装配与启动
    ├── internal/
    │   ├── config, constants, logger, errors
    │   ├── model, repository, service, handler, router
    │   ├── dto, middleware
    ├── database/migrations/        # Schema 管理边界文档
    ├── database/seeds/             # 演示数据说明
    ├── api/openapi.yaml
    └── Dockerfile
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `COMPOSE_PROJECT_NAME` | `cybuildprice` | Compose 英文项目名，保证中文目录也可使用 |
| `DB_NAME` / `DB_USER` / `DB_PASSWORD` | `app` / `app` / `app_pwd` | PostgreSQL 初始化与后端连接信息 |
| `JWT_SECRET` | 示例随机字符串 | 生产环境必须替换为强随机密钥 |
| `FRONTEND_PORT` | `18625` | Nginx 对外端口 |
| `BACKEND_PORT` | `19625` | Gin 对外端口 |
| `DB_PORT` | `5432` | PostgreSQL 本地映射端口 |

## Docker 部署说明

- Compose 顶层 `name: cybuildprice` 与 `.env` 的 `COMPOSE_PROJECT_NAME` 共同避免中文目录名的项目名解析问题；启动命令不需要 `-p`。
- `db_data` 和 `redis_data` 为命名卷，不绑定包含中文字符的宿主机路径；数据会在 `docker compose down` 后保留。如需清理数据，使用 `docker compose down -v`。
- 前端 Nginx 把 `/api/` 代理到 `backend:8080`，前端不硬编码 `localhost`。
- 若端口冲突，修改 `.env` 中 `FRONTEND_PORT`、`BACKEND_PORT` 或 `DB_PORT`，然后执行 `docker compose up -d`。
- 若后端没有变为 healthy，请先运行 `docker compose logs backend`。数据库首次创建完成前，后端会等待 `db` 的健康检查。

## 验证命令

```bash
# 后端依赖、构建、测试
(cd backend && go mod tidy && go build ./... && go test ./...)

# 前端静态构建与类型检查
(cd frontend && npm ci && npm run typecheck && npm run build)

# Compose 配置验证（可在中文目录中执行）
docker compose config --quiet
```

## License

MIT
