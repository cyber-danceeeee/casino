# Casino Go API

本目录为统一 Casino 项目的后端，使用 Go 1.24.13 / Gin / GORM。

完整部署、初始化账号、环境变量、数据库迁移和开发说明见 [项目根目录 README](../README.md)。H5 和管理后台分别位于 `../apps/h5`、`../apps/admin`，统一入口为根目录 `Dockerfile` 和 `compose.yaml`。

```bash
cd server
GOTOOLCHAIN=go1.24.13 go test ./...
GOTOOLCHAIN=go1.24.13 go run ./cmd/casino -c configs/casino.yml
```

直接运行时请导出数据库、Redis 和 MFA 等环境变量，或提供本地私有配置。`configs/casino.yml` 只包含无密钥的默认值。

`deployments/`、`scripts/build/` 和本目录 Makefile 是原工程的独立部署工具，保留作参考；全栈部署请使用根目录 Compose。
