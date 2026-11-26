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

## Nacos 配置中心

支持 Nacos 2.3.2，启动时通过 HTTP OpenAPI 登录并读取 YAML。API、管理员初始化程序和演示数据工具使用同一套读取逻辑。配置中心读取失败、鉴权失败或 YAML 无效时直接停止启动，不回退到本地旧配置，也不在磁盘缓存远程配置。

在私有 `.env` 配置以下变量；账号、密码不要写入仓库：

```dotenv
CASINO_CONFIG_SOURCE=nacos
NACOS_SERVER_ADDR=http://nacos.internal:8848
NACOS_NAMESPACE_ID=
NACOS_GROUP=DEFAULT_GROUP
NACOS_DATA_ID=game
NACOS_USERNAME=casino-reader
NACOS_PASSWORD=填写该账号的密码
NACOS_TIMEOUT_SECONDS=15
```

Nacos 2.3.2 的 `public` 命名空间 ID 留空；自定义命名空间填写实际 ID。`NACOS_SERVER_ADDR` 也可包含 `/nacos` 路径，登录和配置读取合计默认超时 15 秒。当前接入使用 8848 的 HTTP 接口，不依赖 9848。建议项目账号只授予读取该配置的权限。

远程 YAML 采用 `server/configs/casino.yml` 的结构，至少包含 `app`、`http`、`database` 节，HTTP 端口必须有效、数据库驱动为 `mysql`。显式环境变量优先于 Nacos；Docker 中已有的数据库、Redis、RabbitMQ、MFA 密钥覆盖继续生效，与 Compose 内的服务保持一致。数据库迁移容器仍使用 Compose 的数据库连接，不从 Nacos 选择其他数据库。要改用外部数据库，需要同步调整迁移与运行服务的连接及依赖。

配置源 `auto`（默认）在提供 `NACOS_SERVER_ADDR` 时选用 Nacos，否则读取本地 YAML；显式设置 `local` 可用于离线开发。配置修改后重启 API 生效，不对数据库连接池、钱包密钥或供应商运行时做热更新。相对资源路径仍按 `-c` 指定文件所在目录解析；国家、币种与供应商账号继续由后台管理并保存在业务数据库。

只检查读取和格式，不连接业务数据库：

```bash
python3 scripts/dev-api.py --check-config
# Docker 镜像构建后也可以独立检查
docker compose run --rm --no-deps api /app/check-config -c /app/configs/casino.yml
```

`make dev-api` 自动读取根目录私有 `.env`，已导出的进程环境变量优先；启动脚本不执行 `.env` 中的 shell 命令，也不展开其中的变量引用。Nacos 的修改不会由仓库自动发布；更新后按前面的部署命令重建并启动容器。

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

新部署直接使用 `server/database/schema.sql` 和 `seed.sql` 创建单站点数据库，不再执行历史多租户迁移链。`game_site_config` 只允许唯一配置（ID 为 1），业务表没有 `tenant_id`，租户、套餐、域名分发、开站交付和租户账本表均已删除。`site_id=0` 仅表示导入资源目录，`site_id=1` 表示当前 H5 的内容关联。

从上一版本升级时保留原数据卷与 `.env`。安装容器会先完整备份旧库到 `migration_backups` 数据卷，再执行 `server/database/upgrade_legacy.sql`，保留玩家、钱包、订单和管理员记录，删除旧租户结构。包含多个独立站点或租户的数据会拒绝升级，需要先整理为独立数据库。升级不允许把不同品牌的玩家和资金自动合并。

迁移记录包含文件 SHA-256；重复启动跳过已完成安装。MySQL DDL 无法整份事务回滚，失败会留下 dirty 标记并阻止 API 启动。修复或恢复前请保留完整备份；不要通过删数据卷重试。可以用 `docker compose run --rm --entrypoint bash migrate` 检查 `/backups/` 中的备份。

## 多供应商与固定地区

每次部署仍只有一个 H5。在后台 **H5 配置** 设置国家、币种并发布，然后进入 **游戏中心 → 供应商配置**：

1. 新增供应商品牌连接，填写唯一连接编码、接口类型、测试/正式环境和 API 地址。
2. 配置该连接允许的国家、币种、商户凭据和回调 IP 白名单。现有 AG、SoftAggregator 协议可以复用为多个独立品牌账号；新的 API 协议需实现供应商与回调适配器并注册到 `server/internal/game/connections/factory.go`。
3. 将页面上的回调地址登记到供应商平台，测试连接，再同步游戏目录；选择游戏、启用并发布后才会在 H5 展示。

供应商必须明确配置允许的国家和币种；游戏本身的国家/币种限制留空表示继承供应商范围。大厅、搜索、详情、供应商游戏列表、试玩和正式启动统一按**站点已发布的地区**校验，客户端参数和请求头无法切换地区。国家或币种修改后也需发布；已有钱包的币种变更仍要求先完成数据迁移。

连接配置保存后即时生效。密钥使用部署的 `CASINO_MFA_ENCRYPTION_KEY` 派生专用加密密钥，在数据库加密保存，不通过管理 API 回显，也不会进入配置保存请求日志和审计明文。保留 `.env` 中的加密密钥和数据库备份；输入框留空保留已保存凭据。

停用连接或切换地区只限制新展示和启动，不按当前市场阻断已有交易回调。不要在仍有未结算游戏交易时直接更换供应商账号或签名密钥。

原有环境变量接入配置仍可读取；在后台首次保存相同编码后，以数据库配置为准。已有供应商也需要设置明确的允许国家、币种，才会对 H5 展示。

新增的独立账号表和菜单通过 `server/database/updates/` 的增量迁移安装；旧版本已执行的单站点基线保持不变。升级仍使用 `docker compose up -d --build`，重复执行迁移不会覆盖已保存连接。

## 外部服务配置

默认使用本地上传存储，游戏供应商、邮件和 Firebase 均关闭。数据库和队列可正常启动，不依赖原服务器。需要启用时，用 `compose.override.yaml` 为 `api.environment` 增加以下环境变量，密钥从私有 `.env` 引用；不要写入源码或镜像：

| 功能 | 环境变量 |
| --- | --- |
| OSS | `CASINO_STORAGE_DRIVER=oss`、`CASINO_OSS_REGION`、`CASINO_OSS_ENDPOINT`、`CASINO_OSS_BUCKET`、`CASINO_OSS_ACCESS_KEY_ID`、`CASINO_OSS_ACCESS_KEY_SECRET` |
| FCM | `CASINO_FCM_ENABLED=true`、`CASINO_FCM_PROJECT_ID`、`CASINO_FCM_CREDENTIALS_FILE`，并将服务账号 JSON 只读挂载到对应容器路径 |
| SoftAggregator | `GAME_SOFTAGGREGATOR_ENABLED`、`GAME_SOFTAGGREGATOR_API_LOGIN`、`GAME_SOFTAGGREGATOR_API_PASSWORD`、`GAME_SOFTAGGREGATOR_SALT_KEY`，其他参数见 `server/.env.example` |
| AG | `GAME_AG_ENABLED=true`、`GAME_AG_MERCHANT_CODE`、`GAME_AG_SECRET_KEY`、`GAME_AG_API_BASE_URL` |

只加入 `.env` 不会自动将这些可选变量传入 API，必须在 override 的 `environment` 中显式引用。上传目录 `/app/static` 的命名卷首次创建时会复制镜像内的 H5 媒体；已有卷升级不会覆盖文件，新增内置媒体需同步到该卷并保留用户上传。

## 本地开发与验证

前端使用 Node.js 22.12+（推荐 Node 22 最新维护版），后端固定 Go 1.24.13。可使用 `GOTOOLCHAIN=go1.24.13` 自动下载对应工具链。

```bash
make install
make check
# 分别在三个终端中启动；API 从私有 .env 选择 Nacos 或本地 YAML
make dev-api
make dev-h5
make dev-admin
```

本地 Vite 将 `/api`、`/static` 代理到 `127.0.0.1:1639`。Go 通过 `CASINO_MYSQL_DSN`、`CASINO_REDIS_DSN`、`CASINO_MFA_ENCRYPTION_KEY` 等环境变量覆盖所选配置源；根目录 `.env` 由 Compose 和 `make dev-api` 读取，直接运行 Go 时需要自行导出对应环境变量。

部署后检查（端口可通过环境变量覆盖）：

```bash
python3 scripts/smoke-test.py
docker compose run --rm migrate
```

GitHub Actions 会验证两个前端构建、单站点后台检查、Go 测试、全新 Docker 部署、管理员 MFA 登录、H5 草稿发布与迁移重放。端到端测试仅在临时数据库运行，会修改临时管理员密码和 H5 测试配置。服务器部署使用上述 Compose 命令，仓库推送不会自动修改服务器。
