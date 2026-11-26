# Casino

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
- 平台后台：`http://服务器IP:8080/platform/login`（`ADMIN_PORT=8080`）
- 租户后台：`http://服务器IP:8080/tenant/login`
- 就绪检查：两个入口下的 `/health/ready`

首次平台账号为 `system-admin`，密码读取服务器 `.env` 中的 `CASINO_BOOTSTRAP_ADMIN_PASSWORD`。按登录页面提示绑定验证器 MFA，再修改初始密码。初始化程序只为密码为空的预置管理员设置密码；重启和升级不会覆盖已有密码。

Compose 会依次启动 MySQL、Redis、RabbitMQ，执行数据库迁移、初始化管理员、启动 API，最后开放前端。`migrate`、`bootstrap` 正常完成后显示 `Exited (0)`。数据库、缓存、队列、上传目录均使用命名卷；只对外开放两个前端端口。

新库没有旧服务器的租户和业务数据。登录平台后台后创建品牌/租户、配置站点域名并发布配置，H5 才能取得对应站点数据。`server/cmd/seed-h5` 是原有演示数据工具，不会在生产部署时自动执行。

## 域名和 HTTPS

可将 H5 和后台的不同域名分别反向代理到本机两个端口，在外层 Nginx/Caddy/负载均衡器配置 TLS 并保留原始 `Host`。站点按域名识别租户，应将相同域名录入后台站点配置。两个 Vue 应用都部署在各自站点根路径，不要把后台直接挂到 `/admin/` 子路径。

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

此流程面向**全新数据库**。已有数据库没有迁移记录，不能直接接入后重跑历史 SQL；请先在备份库核对表结构和迁移基线。旧工程的明文生产凭据没有纳入本仓库历史，需要迁移原数据或外部资源时另行配置。

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

GitHub Actions 会验证两个前端构建、后台现有检查、Go 测试、全新 Docker 部署以及迁移重放。服务器部署使用上述 Compose 命令，仓库推送不会自动修改服务器。
