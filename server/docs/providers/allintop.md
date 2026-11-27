# AllInTop v2（GoldenGateX 测试账号）

适配器编码为 `allintop`。本次账号仅支持印度尼西亚 `ID / IDR`、单一钱包（Seamless）。`IDRK` 是另一种带倍率的币种，不可作为 IDR 使用。账号密码用于登录供应商后台；API 使用后台个人资料页的 `clientId` / `clientSecret`，无需填写后台密码或 agentCode。

## Nacos

在 public（Namespace ID 空）、DEFAULT_GROUP、DataId `game` 的现有 YAML **game 节点下**合并以下内容，保留现有其他供应商配置。不要再创建第二个顶层 `game` 节点。真实凭据只存 Nacos，不提交 Git。

```yaml
game:
  allInTop:
    # 当前测试账号这两个品牌的目录返回上游错误 500；恢复后可移除排除项。
    excludedVendorCodes: ["board-poker", "mini-spribe"]
    enabled: true
    apiBaseURL: "https://dcyqv8f2id.com/api/v2"
    clientId: "填写供应商后台的 clientId"
    clientSecret: "填写供应商后台的 clientSecret"
    callbackIPAllowlist: []
    requestTimeoutSeconds: 15
    connectTimeoutSeconds: 3
    maxTransactionAmountMinor: 100000000
```

金额的内部最小单位固定为 0.01 IDR。上例单次交易上限为 1,000,000 IDR，按实际账号限额调整。回调金额必须能精确表示为两位小数，不做汇率换算或四舍五入。

Nacos 在 API 进程启动时读取；修改后在服务器运行 `docker compose up -d --no-deps --force-recreate --wait api`。后台“供应商配置”保存的同编码账号优先于 Nacos；使用 Nacos 管理此账号时不要另存数据库覆盖配置。

## 回调登记

供应商“Profile / Callback URL”登记的是**基地址**，不能填写泛化的 `/game/providers` 列表路径。以当前测试服务器为例：

```text
http://64.83.34.122:18081/api/v1/game/providers/allintop/callback
```

供应商会追加固定路径，对应完整接口为：

```text
POST /api/v1/game/providers/allintop/callback/api/balance
POST /api/v1/game/providers/allintop/callback/api/transaction
POST /api/v1/game/providers/allintop/callback/api/batch-transactions
```

Basic 认证用户名为 `clientId`、密码为 `clientSecret`。正式环境必须使用 HTTPS 域名以保护认证信息；收到供应商明确提供的回调出口 IP 后，可填写 `callbackIPAllowlist`（IP 或 CIDR）。不能把我方服务器 IP 当作供应商回调来源白名单。

## 行为与验证

- 缓存复用 token，并合并并发认证；失败后冷却 30 秒，不自动循环重试 401、403、429。供应商创建 token 限制为每 30 秒 5 次。
- 从 `/vendors/list` 获取品牌，再逐个调用 `/games/list`（默认 en，印尼语只用于文档明确支持的品牌）。`excludedVendorCodes` 仅用于 Nacos 管理的默认 allintop 账号，用来显式排除供应商尚不可用的品牌；移除排除项并重启后即可再次获取。任一品牌请求失败或总目录为空则整次失败，避免部分目录导致已有游戏被误下架。
- 目录内部标识为 `vendorCode/gameCode`，启动拆分调用 `/game/launch-url`。单一钱包用户由启动创建；不调用转账钱包的建用户或充值接口。未公布试玩接口，返回不支持。
- 回调负数扣款、正数派奖、零金额允许结束回合；`transactionCode` 跨操作及玩家去重，重复返回错误 6，已结束回合的后续正常交易返回 7。
- 取消忽略其 amount，只逐笔冲销该 historyId 下此前成功的账本记录；支持结算后取消、取消先于下注到达。取消与批量交易都是数据库原子事务。
- 批量交易必须同一玩家，任一失败全部回滚，服务执行设置 4.5 秒截止时间。
- 供应商初始地区规则为 ID / IDR。只有当前启用国家为印度尼西亚且游戏在后台启用后，H5 才会展示和允许启动；不会将现有 PKR 钱包自动换成 IDR。
- 不调用 RTP、Autocall、转账或下注接口。上线前仍需供应商配合完成真实取消、派彩与重试联调。

文档：[供应商 API 文档](https://samyarieiv636.github.io/topapidoc.github.io/)。文档密码不存仓库。
