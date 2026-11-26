# Casino 单站点版

一个管理后台直接配置一个 H5，包含玩家、资金、游戏、活动、风控、站点配置和管理员权限管理。无需创建租户、选择套餐或切换平台/租户登录入口。

H5、管理后台与 Go API 的统一仓库。通过根目录 Dockerfile 和 Docker Compose 一起部署；服务器只需要 Docker Engine 和 Compose v2，构建时会安装 Node/Go 依赖。建议服务器至少 4 GB 内存并预留构建磁盘空间；内存较小的机器可在其他机器构建镜像后再部署。

| 目录 | 用途 |
| --- | --- |
| `apps/h5` | Vue 3 / Vite 玩家 H5 |
| `apps/admin` | Vue 3 / Element Plus 管理后台（原 game_casino_backend） |
| `server` | Go 1.24.13 / Gin API、数据库模型和业务代码 |
| `deploy` | Nginx、迁移顺序清单和迁移脚本 |
| `scripts` | 环境初始化、部署冒烟检查 |

## 服务器首次部署

```bash
git clone https://github.com/cyber-danceeeee/casino.git
cd casino
python3 scripts/init-env.py
docker compose up -d --build --wait --wait-timeout 600
docker compose ps -a
```

初始化脚本只创建一次 `.env`，自动生成数据库、RabbitMQ、MFA 和初始管理员密码，权限为 `600`。也可以手动复制 `.env.example` 到 `.env` 并填写空值。保留 `.env`，不要在升级时重新生成；MFA 加密密钥改变会导致已有 MFA 密文无法解密。

默认访问：

- H5：`http://服务器IP/`（`H5_PORT=80`）
- 管理后台：`http://服务器IP:8080/login`（`ADMIN_PORT=8080`）；旧登录地址会跳转到此统一入口
- 就绪检查：两个入口下的 `/health/ready`

首次管理员账号为 `system-admin`，密码读取服务器 `.env` 中的 `CASINO_BOOTSTRAP_ADMIN_PASSWORD`。按登录页面提示绑定验证器 MFA，再修改初始密码。初始化程序只为密码为空的预置管理员设置密码；重启和升级不会覆盖已有密码。

Compose 会依次启动 MySQL、Redis、RabbitMQ，执行数据库迁移、初始化唯一站点和管理员、启动 API，最后开放前端。`migrate`、`bootstrap` 正常完成后显示 `Exited (0)`。数据库、缓存、队列、上传目录均使用命名卷；只对外开放两个前端端口。

新库首次启动会自动创建唯一站点，按已启用的默认国家配置带出币种、语言和时区，并初始化大厅与运营配置。登录后台后直接进入 **H5 配置**：填写 H5 访问地址、站点名称、Logo、主题、国家、语言、客服、SEO 和维护状态，保存草稿后点击发布。国家选择会带出对应币种；已存在玩家钱包时不允许直接更换站点币种，需要先迁移余额及相关业务数据。

H5 地址接受 `https://h5.example.com` 或本地 `http://服务器IP:端口`，无需再登记租户域名或传递租户请求头。H5 始终读取当前安装绑定的唯一站点；后台预览入口打开配置的 H5 地址。页面标题、Logo、主题主按钮和背景会应用已发布配置。`server/cmd/seed-h5` 是原有演示数据工具，不会自动插入演示玩家或支付渠道。

## 域名和 HTTPS

可将 H5 和后台的不同域名分别反向代理到本机两个端口，在外层 Nginx/Caddy/负载均衡器配置 TLS 并保留原始 `Host`。将 H5 对外访问地址填写到后台 H5 配置中即可；API 不按域名分流站点。两个 Vue 应用都部署在各自站点根路径，不要把后台直接挂到 `/admin/` 子路径。

如果外层代理也在这台服务器上，修改 `.env` 的 `BIND_ADDRESS=127.0.0.1`，并按需调整 `H5_PORT` / `ADMIN_PORT`。容器内 Nginx 已处理 SPA 页面刷新、`/api/` 和 `/static/` 转发、100 MB 上传。公网使用前应配置 HTTPS。

## 更新、日志与备份

```bash
git pull --ff-only origin main
docker compose up -d --build --wait --wait-timeout 600
docker compose logs -f --tail=100 api web
docker compose down
```

`down` 保留数据卷；`down -v` 会删除全部部署数据。升级前备份 MySQL、`uploads` 卷和 `.env`。例如备份数据库（文件包含业务数据，应妥善保存）：

```bash
docker compose exec -T mysql sh -c 'MYSQL_PWD="$MYSQL_PASSWORD" mysqldump -u casino --single-transaction --no-tablespaces game_casino' > casino-backup.sql
```

迁移按 `deploy/migrations.txt` 的依赖顺序执行，成功后记录文件 SHA-256。重复启动跳过已完成迁移；中途失败或文件被修改会阻止 API 启动，避免忽略错误。MySQL DDL 不支持整份文件回滚：出现 dirty 记录时先备份并人工检查失败语句，恢复数据库或修复后再处理迁移记录。新增迁移应追加清单，不要修改已执行文件。

从本仓库上一版本升级时，保留原数据卷和 `.env`，新增迁移会创建单站点绑定表。只有一个已有站点的数据库会沿用其 ID、配置和业务数据，重启不会再创建站点。旧 SaaS 的租户管理、套餐和开站接口在此版本关闭，管理员权限控制与 MFA 保留。

包含多个未删除站点/租户、重复管理员用户名或没有站点归属的旧玩家数据时，初始化会停止并输出原因，不会自动混合独立品牌、玩家和资金记录。请先备份并将需要运行的唯一站点整理到独立数据库，再升级。历史 `tenant_id` / `site_id` 仅作为内部存储关联字段保留，客户端无法选择站点，旧表和已有迁移不会删除。

**直接接入外部旧库**若没有迁移记录，不能重跑历史 SQL；请先在备份库核对表结构和迁移基线。旧工程的明文生产凭据没有纳入本仓库历史，需要迁移原数据或外部资源时另行配置。

## 外部服务配置

默认使用本地上传存储，游戏供应商、邮件和 Firebase 均关闭。数据库和队列可正常启动，不依赖原服务器。需要启用时，用 `compose.override.yaml` 为 `api.environment` 增加以下环境变量，密钥从私有 `.env` 引用；不要写入源码或镜像：

| 功能 | 环境变量 |
| --- | --- |
| OSS | `CASINO_STORAGE_DRIVER=oss`、`CASINO_OSS_REGION`、`CASINO_OSS_ENDPOINT`、`CASINO_OSS_BUCKET`、`CASINO_OSS_ACCESS_KEY_ID`、`CASINO_OSS_ACCESS_KEY_SECRET` |
| AWS SES | `CASINO_EMAIL_ENABLED=true`、`CASINO_EMAIL_PROVIDER=AWS_SES`、`CASINO_EMAIL_FROM`、`CASINO_EMAIL_REGION`、`CASINO_EMAIL_ACCESS_KEY_ID`、`CASINO_EMAIL_SECRET_ACCESS_KEY` |
| Resend | `CASINO_EMAIL_ENABLED=true`、`CASINO_EMAIL_PROVIDER=RESEND`、`CASINO_EMAIL_FROM`、`CASINO_RESEND_API_KEY` |
| FCM | `CASINO_FCM_ENABLED=true`、`CASINO_FCM_PROJECT_ID`、`CASINO_FCM_CREDENTIALS_FILE`，并将服务账号 JSON 只读挂载到对应容器路径 |
| SoftAggregator | `GAME_SOFTAGGREGATOR_ENABLED`、`GAME_SOFTAGGREGATOR_API_LOGIN`、`GAME_SOFTAGGREGATOR_API_PASSWORD`、`GAME_SOFTAGGREGATOR_SALT_KEY`，其他参数见 `server/.env.example` |
| AG | `GAME_AG_ENABLED=true`、`GAME_AG_MERCHANT_CODE`、`GAME_AG_SECRET_KEY`、`GAME_AG_API_BASE_URL` |

只加入 `.env` 不会自动将这些可选变量传入 API，必须在 override 的 `environment` 中显式引用。上传目录 `/app/static` 的命名卷首次创建时会复制镜像内的 H5 媒体；已有卷升级不会覆盖文件，新增内置媒体需同步到该卷并保留用户上传。

## 本地开发与验证

前端使用 Node.js 22.12+（推荐 Node 22 最新维护版），后端固定 Go 1.24.13。可使用 `GOTOOLCHAIN=go1.24.13` 自动下载对应工具链。

```bash
make install
make check
# 分别在三个终端中启动；API 的数据库/Redis需配置到 server/configs/casino.yml
make dev-api
make dev-h5
make dev-admin
```

本地 Vite 将 `/api`、`/static` 代理到 `127.0.0.1:1639`。Go 通过 `CASINO_MYSQL_DSN`、`CASINO_REDIS_DSN`、`CASINO_MFA_ENCRYPTION_KEY` 等环境变量覆盖 YAML；根目录 `.env` 由 Compose 读取，直接运行 Go 时需要自行导出对应环境变量。

部署后检查（端口可通过环境变量覆盖）：

```bash
python3 scripts/smoke-test.py
docker compose run --rm migrate
```

GitHub Actions 会验证两个前端构建、单站点后台检查、Go 测试、全新 Docker 部署、管理员 MFA 登录、H5 草稿发布与迁移重放。端到端测试仅在临时数据库运行，会修改临时管理员密码和 H5 测试配置。服务器部署使用上述 Compose 命令，仓库推送不会自动修改服务器。
