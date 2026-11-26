export type Tone = "success" | "warning" | "danger" | "info" | "neutral";

export type Metric = {
  label: string;
  value: string;
  trend: string;
  tone: Tone;
};

export type TableColumn = {
  key: string;
  label: string;
  align?: "left" | "right" | "center";
};

export type TableRow = {
  statusTone?: Tone;
  [key: string]: string | undefined;
};

export type ModuleConfig = {
  key: string;
  title: string;
  section: string;
  description: string;
  columns: TableColumn[];
};

const userColumns: TableColumn[] = [
  { key: "account", label: "账号" },
  { key: "phone", label: "手机号" },
  { key: "country", label: "国家" },
  { key: "currency", label: "币种" },
  { key: "vip", label: "VIP" },
  { key: "balance", label: "余额", align: "right" },
  { key: "status", label: "状态", align: "center" },
  { key: "updatedAt", label: "更新时间" },
];

const financeColumns: TableColumn[] = [
  { key: "orderNo", label: "单号" },
  { key: "user", label: "用户" },
  { key: "type", label: "类型" },
  { key: "amount", label: "金额", align: "right" },
  { key: "currency", label: "币种" },
  { key: "reference", label: "关联单据" },
  { key: "status", label: "状态", align: "center" },
  { key: "createdAt", label: "创建时间" },
];

const gameColumns: TableColumn[] = [
  { key: "gameCode", label: "游戏编码" },
  { key: "name", label: "名称" },
  { key: "provider", label: "Provider" },
  { key: "category", label: "平台分类" },
  { key: "providerStatus", label: "Provider状态", align: "center" },
  { key: "platformStatus", label: "平台状态", align: "center" },
  { key: "flags", label: "标签" },
  { key: "updatedAt", label: "更新时间" },
];

const betColumns: TableColumn[] = [
  { key: "orderNo", label: "平台订单号" },
  { key: "providerTx", label: "Provider交易号" },
  { key: "roundNo", label: "Round" },
  { key: "user", label: "用户" },
  { key: "game", label: "游戏" },
  { key: "bet", label: "投注", align: "right" },
  { key: "win", label: "派奖", align: "right" },
  { key: "status", label: "状态", align: "center" },
];

const basicColumns: TableColumn[] = [
  { key: "name", label: "名称" },
  { key: "owner", label: "负责人" },
  { key: "scope", label: "范围" },
  { key: "status", label: "状态", align: "center" },
  { key: "updatedAt", label: "更新时间" },
];

const moduleDefinitions = [
  [
    "user-list",
    "用户列表",
    "用户中心",
    "管理用户账户、登录状态、VIP、KYC、标签和资金入口。",
    userColumns,
  ],
  [
    "login-records",
    "登录记录",
    "用户中心",
    "追踪登录 IP、设备指纹、地理位置和异常登录。",
    userColumns,
  ],
  [
    "device-management",
    "设备管理",
    "用户中心",
    "管理设备指纹、同设备账号、可信设备和封禁设备。",
    userColumns,
  ],
  [
    "risk-users",
    "风险用户",
    "用户中心",
    "集中处理多账户、奖金套利、高频提现等风险用户。",
    userColumns,
  ],

  [
    "wallet-accounts",
    "钱包账户",
    "资金中心",
    "查看现金、奖金、可提现、冻结余额和钱包状态。",
    financeColumns,
  ],
  [
    "wallet-ledgers",
    "钱包流水",
    "资金中心",
    "所有资金变化不可修改、不可删除、可审计。",
    financeColumns,
  ],
  [
    "manual-adjustments",
    "人工加减款",
    "资金中心",
    "通过审批流执行人工加款或扣款并生成流水。",
    financeColumns,
  ],
  [
    "deposit-orders",
    "充值订单",
    "资金中心",
    "处理支付订单、补单、回调、签名和渠道成功率。",
    financeColumns,
  ],
  [
    "withdrawal-orders",
    "提现订单",
    "资金中心",
    "风控审核、财务审核、拒绝原因和冻结解冻。",
    financeColumns,
  ],
  [
    "financial-reconciliation",
    "财务对账",
    "资金中心",
    "对账支付渠道、钱包流水、Provider 账单和异常差异。",
    financeColumns,
  ],
  [
    "deposit-methods",
    "充值方式",
    "资金中心",
    "配置 H5 可用的充值方式、图标、金额范围和排序。",
    financeColumns,
  ],
  [
    "deposit-channels",
    "充值渠道",
    "资金中心",
    "配置充值方式下的支付渠道、费率和金额限制。",
    financeColumns,
  ],
  [
    "deposit-amount-options",
    "充值金额档位",
    "资金中心",
    "配置 H5 充值金额选项和默认奖励。",
    financeColumns,
  ],

  [
    "game-sync",
    "游戏厂商",
    "游戏中心",
    "维护游戏厂商、连接状态与游戏同步任务。",
    gameColumns,
  ],
  [
    "game-list",
    "游戏列表",
    "游戏中心",
    "统一游戏目录，支持上下架、分类、标签、排序和限制。",
    gameColumns,
  ],
  [
    "game-categories",
    "游戏分类",
    "游戏中心",
    "维护平台分类与 Provider 分类映射。",
    gameColumns,
  ],
  [
    "game-launch-records",
    "游戏启动记录",
    "游戏中心",
    "追踪游戏启动、launchUrl、失败原因和用户设备。",
    gameColumns,
  ],

  [
    "bet-records",
    "投注记录",
    "投注中心",
    "查看投注、派奖、GGR、状态、回滚和重复回调。",
    betColumns,
  ],
  [
    "rollback-records",
    "回滚记录",
    "投注中心",
    "审查 Provider 回滚、退款和钱包补偿。",
    betColumns,
  ],
  [
    "abnormal-bets",
    "异常投注",
    "投注中心",
    "发现重复回调、签名失败、异常派奖和高风险投注。",
    betColumns,
  ],
  [
    "provider-callback-logs",
    "Provider回调日志",
    "投注中心",
    "保留 Provider 原始请求响应、签名结果和耗时。",
    betColumns,
  ],

  [
    "banners",
    "Banner",
    "运营中心",
    "配置 H5 首页轮播、弹窗、跳转和展示周期。",
    basicColumns,
  ],
  [
    "announcements",
    "公告",
    "运营中心",
    "维护站内公告、跑马灯、弹窗通知和多语言内容。",
    basicColumns,
  ],
  [
    "activities",
    "活动",
    "运营中心",
    "管理首充、签到、转盘、红包雨、排行榜等活动。",
    basicColumns,
  ],
  [
    "vip",
    "VIP",
    "运营中心",
    "配置 VIP 等级、保级、升级、权益和奖励。",
    basicColumns,
  ],
  [
    "push-messages",
    "推送消息",
    "运营中心",
    "定向推送活动、提现提醒、回流消息和站内信。",
    basicColumns,
  ],
  [
    "popups",
    "弹窗投放",
    "运营中心",
    "选择平台登记的 H5 弹窗组件并配置投放顺序、跳转和受众规则。",
    basicColumns,
  ],
  [
    "music",
    "音乐管理",
    "运营中心",
    "管理大厅、登录页、加载页及活动使用的背景音乐。",
    basicColumns,
  ],
  [
    "social-links",
    "社媒管理",
    "运营中心",
    "管理分享渠道、社交媒体图标、链接和转发文案。",
    basicColumns,
  ],
  [
    "h5-navigation",
    "H5 菜单管理",
    "运营中心",
    "配置 H5 底部菜单、图标、文本以及页面或游戏跳转。",
    basicColumns,
  ],
  [
    "vouchers",
    "优惠券管理",
    "运营中心",
    "配置充值优惠券、资格规则、奖励档位和使用限制。",
    basicColumns,
  ],

  [
    "risk-dashboard",
    "风控看板",
    "风控中心",
    "汇总风险事件、审核、封禁及异常趋势。",
    basicColumns,
  ],
  [
    "risk-rules",
    "风控规则",
    "风控中心",
    "维护结构化风险条件、动作、范围与测试记录。",
    basicColumns,
  ],
  [
    "risk-events",
    "风险事件",
    "风控中心",
    "查看自动命中和人工标记的风险事件。",
    basicColumns,
  ],
  [
    "review-queue",
    "审核队列",
    "风控中心",
    "处理提现、人工加减款、风险用户和大额中奖审核。",
    basicColumns,
  ],
  [
    "user-associations",
    "用户关联",
    "风控中心",
    "分析设备、IP、指纹、推荐关系及收款账户关联。",
    basicColumns,
  ],
  [
    "blacklist",
    "黑名单管理",
    "风控中心",
    "维护手机号、银行卡、设备、IP、支付账户黑名单。",
    basicColumns,
  ],
  [
    "withdrawal-risk",
    "提现风控审核",
    "风控中心",
    "复用资金中心提现工作流完成风险判断。",
    financeColumns,
  ],
  [
    "risk-operation-logs",
    "风控操作日志",
    "风控中心",
    "查看不可修改的风控操作审计记录。",
    basicColumns,
  ],

  [
    "agent-list",
    "代理列表",
    "代理中心",
    "管理代理账号、层级、钱包和归因数据。",
    basicColumns,
  ],
  [
    "agent-hierarchy",
    "代理层级",
    "代理中心",
    "查看和维护代理父子层级。",
    basicColumns,
  ],
  [
    "channel-management",
    "渠道管理",
    "代理中心",
    "维护渠道码、落地页、注册归因和首充归因。",
    basicColumns,
  ],
  [
    "commission-plans",
    "佣金方案",
    "代理中心",
    "配置 CPA、Revenue Share、混合和阶梯佣金。",
    basicColumns,
  ],
  [
    "commission-records",
    "佣金记录",
    "代理中心",
    "查看不可变佣金计算记录。",
    financeColumns,
  ],
  [
    "agent-settlements",
    "代理结算",
    "代理中心",
    "生成、审核和支付代理结算。",
    financeColumns,
  ],
  [
    "agent-wallets",
    "代理钱包",
    "代理中心",
    "查看代理余额及不可变账本。",
    financeColumns,
  ],
  [
    "agent-operation-logs",
    "代理操作日志",
    "代理中心",
    "查看不可修改的代理中心审计记录。",
    basicColumns,
  ],

  [
    "financial-report",
    "财务报表",
    "报表中心",
    "统计充值、提现、GGR、NGR、Bonus 成本和支付费用。",
    financeColumns,
  ],
  [
    "game-report",
    "游戏报表",
    "报表中心",
    "按游戏统计玩家、投注、派奖、实际 RTP 和 GGR。",
    gameColumns,
  ],
  [
    "provider-report",
    "Provider报表",
    "报表中心",
    "按 Provider 统计 GGR、错误率、回调延迟和局数。",
    gameColumns,
  ],
  [
    "user-report",
    "用户报表",
    "报表中心",
    "统计用户充值、提现、投注、派奖、净输赢和活跃。",
    userColumns,
  ],
  [
    "channel-report",
    "渠道报表",
    "报表中心",
    "分析渠道注册、首充、留存、ROI 和佣金成本。",
    basicColumns,
  ],
  [
    "agent-report",
    "代理报表",
    "报表中心",
    "统计代理归因用户、收入、佣金和结算。",
    financeColumns,
  ],
  [
    "promotion-report",
    "活动报表",
    "报表中心",
    "分析活动参与、奖励成本、流水完成度和 ROI。",
    financeColumns,
  ],
  [
    "export-tasks",
    "导出任务",
    "报表中心",
    "管理异步生成的 CSV 和 XLSX 报表文件。",
    basicColumns,
  ],

  [
    "admins",
    "管理员管理",
    "系统管理",
    "管理后台账号、MFA、数据权限和登录安全。",
    basicColumns,
  ],
  [
    "roles",
    "角色管理",
    "系统管理",
    "维护管理员角色、数据范围、权限和菜单。",
    basicColumns,
  ],
  [
    "permissions",
    "权限管理",
    "系统管理",
    "维护 API、页面、菜单、动作及数据权限。",
    basicColumns,
  ],
  [
    "departments",
    "部门管理",
    "系统管理",
    "维护后台组织层级和管理员归属。",
    basicColumns,
  ],
  [
    "menus",
    "菜单管理",
    "系统管理",
    "维护后台目录、菜单、按钮和外部链接。",
    basicColumns,
  ],
  [
    "system-config",
    "系统配置",
    "系统管理",
    "配置币种、国家、限额、开关和 Provider 密钥占位。",
    basicColumns,
  ],
  [
    "audit-logs",
    "审计日志",
    "系统管理",
    "审计谁在何时从哪里修改了什么。",
    basicColumns,
  ],
  [
    "admin-login-logs",
    "登录日志",
    "系统管理",
    "查看管理员成功和失败的登录记录。",
    basicColumns,
  ],
  [
    "scheduled-tasks",
    "定时任务",
    "系统管理",
    "管理注册过的安全内部定时任务。",
    basicColumns,
  ],
  [
    "system-monitor",
    "系统监控",
    "系统管理",
    "监控 MySQL、Redis、Provider API、回调成功率和队列积压。",
    basicColumns,
  ],
] as const;

const buildConfig = (
  definition: (typeof moduleDefinitions)[number],
): ModuleConfig => {
  const [key, title, section, description, columns] = definition;
  return {
    key,
    title,
    section,
    description,
    columns,
  };
};

export const moduleConfigs: Record<string, ModuleConfig> = Object.fromEntries(
  moduleDefinitions.map((definition) => {
    const config = buildConfig(definition);
    return [config.key, config];
  }),
);
