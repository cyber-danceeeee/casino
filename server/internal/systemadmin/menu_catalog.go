package systemadmin

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type menuCatalogItem struct {
	Code   string
	NameZh string
	NameEn string
	Path   string
}

type menuCatalogGroup struct {
	Code     string
	NameZh   string
	NameEn   string
	Path     string
	Sort     int
	Children []menuCatalogItem
}

var adminMenuCatalog = []menuCatalogGroup{
	{Code: "dashboard", NameZh: "数据看板", NameEn: "Dashboard", Path: "/", Sort: 10},
	{Code: "users", NameZh: "用户中心", NameEn: "User Center", Sort: 100, Children: []menuCatalogItem{
		{Code: "user-list", NameZh: "用户列表", NameEn: "User List", Path: "/users/list"},
		{Code: "login-records", NameZh: "登录记录", NameEn: "Login Records", Path: "/users/login-records"},
		{Code: "device-management", NameZh: "设备管理", NameEn: "Device Management", Path: "/users/devices"},
		{Code: "risk-users", NameZh: "风险用户", NameEn: "Risk Users", Path: "/users/risk"},
	}},
	{Code: "funds", NameZh: "资金中心", NameEn: "Funds Center", Sort: 200, Children: []menuCatalogItem{
		{Code: "wallet-accounts", NameZh: "钱包账户", NameEn: "Wallet Accounts", Path: "/funds/wallets"},
		{Code: "wallet-ledgers", NameZh: "钱包流水", NameEn: "Wallet Ledgers", Path: "/funds/ledgers"},
		{Code: "manual-adjustments", NameZh: "人工加减款", NameEn: "Manual Adjustments", Path: "/funds/manual-adjustments"},
		{Code: "deposit-orders", NameZh: "充值订单", NameEn: "Deposit Orders", Path: "/funds/deposits"},
		{Code: "withdrawal-orders", NameZh: "提现订单", NameEn: "Withdrawal Orders", Path: "/funds/withdrawals"},
		{Code: "financial-reconciliation", NameZh: "财务对账", NameEn: "Financial Reconciliation", Path: "/funds/reconciliation"},
		{Code: "deposit-methods", NameZh: "充值方式", NameEn: "Deposit Methods", Path: "/funds/deposit-methods"},
		{Code: "deposit-channels", NameZh: "充值渠道", NameEn: "Deposit Channels", Path: "/funds/deposit-channels"},
		{Code: "deposit-amount-options", NameZh: "充值金额档位", NameEn: "Deposit Amount Options", Path: "/funds/deposit-amount-options"},
	}},
	{Code: "games", NameZh: "游戏中心", NameEn: "Game Center", Sort: 300, Children: []menuCatalogItem{
		{Code: "game-sync", NameZh: "游戏厂商", NameEn: "Game Providers", Path: "/games/sync"},
		{Code: "game-list", NameZh: "游戏列表", NameEn: "Game List", Path: "/games/list"},
		{Code: "game-categories", NameZh: "游戏分类", NameEn: "Game Categories", Path: "/games/categories"},
		{Code: "lobby-sections", NameZh: "大厅配置", NameEn: "Lobby Sections", Path: "/games/lobby-sections"},
		{Code: "game-launch-records", NameZh: "游戏启动记录", NameEn: "Game Launch Records", Path: "/games/launch-records"},
	}},
	{Code: "bets", NameZh: "投注中心", NameEn: "Betting Center", Sort: 400, Children: []menuCatalogItem{
		{Code: "bet-records", NameZh: "投注记录", NameEn: "Betting Records", Path: "/bets/records"},
		{Code: "rollback-records", NameZh: "回滚记录", NameEn: "Rollback Records", Path: "/bets/rollbacks"},
		{Code: "abnormal-bets", NameZh: "异常投注", NameEn: "Abnormal Bets", Path: "/bets/abnormal"},
		{Code: "provider-callback-logs", NameZh: "Provider回调日志", NameEn: "Provider Callback Logs", Path: "/bets/provider-callbacks"},
		{Code: "provider-reconciliation", NameZh: "Provider对账", NameEn: "Provider Reconciliation", Path: "/bets/provider-reconciliation"},
	}},
	{Code: "operations", NameZh: "运营中心", NameEn: "Operations Center", Sort: 500, Children: []menuCatalogItem{
		{Code: "asset-library", NameZh: "素材中心", NameEn: "Asset Library", Path: "/operations/assets"},
		{Code: "banners", NameZh: "Banner", NameEn: "Banners", Path: "/operations/banners"},
		{Code: "announcements", NameZh: "公告", NameEn: "Announcements", Path: "/operations/announcements"},
		{Code: "activities", NameZh: "活动", NameEn: "Promotions", Path: "/operations/activities"},
		{Code: "vip", NameZh: "VIP", NameEn: "VIP", Path: "/operations/vip"},
		{Code: "push-messages", NameZh: "推送消息", NameEn: "Push Notifications", Path: "/operations/push-messages"},
		{Code: "popups", NameZh: "弹窗投放", NameEn: "Popup Campaigns", Path: "/operations/popups"},
		{Code: "music", NameZh: "音乐管理", NameEn: "Music Management", Path: "/operations/music"},
		{Code: "social-links", NameZh: "社媒管理", NameEn: "Social Links", Path: "/operations/social-links"},
		{Code: "h5-navigation", NameZh: "H5 菜单管理", NameEn: "H5 Navigation", Path: "/operations/h5-navigation"},
		{Code: "vouchers", NameZh: "优惠券管理", NameEn: "Voucher Management", Path: "/operations/vouchers"},
	}},
	{Code: "risk", NameZh: "风控中心", NameEn: "Risk Center", Sort: 600, Children: []menuCatalogItem{
		{Code: "risk-dashboard", NameZh: "风控看板", NameEn: "Risk Dashboard", Path: "/risk/dashboard"},
		{Code: "risk-rules", NameZh: "风控规则", NameEn: "Risk Rules", Path: "/risk/rules"},
		{Code: "risk-events", NameZh: "风险事件", NameEn: "Risk Events", Path: "/risk/events"},
		{Code: "review-queue", NameZh: "审核队列", NameEn: "Review Queue", Path: "/risk/review-queue"},
		{Code: "user-associations", NameZh: "用户关联", NameEn: "User Associations", Path: "/risk/associations"},
		{Code: "blacklist", NameZh: "黑名单管理", NameEn: "Blacklist", Path: "/risk/blacklist"},
		{Code: "withdrawal-risk", NameZh: "提现风控审核", NameEn: "Withdrawal Review", Path: "/risk/withdrawals"},
		{Code: "risk-operation-logs", NameZh: "风控操作日志", NameEn: "Risk Operation Logs", Path: "/risk/operation-logs"},
	}},
	{Code: "agents", NameZh: "代理中心", NameEn: "Agent Center", Sort: 700, Children: []menuCatalogItem{
		{Code: "agent-list", NameZh: "代理列表", NameEn: "Agent List", Path: "/agents/list"},
		{Code: "agent-hierarchy", NameZh: "代理层级", NameEn: "Agent Hierarchy", Path: "/agents/hierarchy"},
		{Code: "channel-management", NameZh: "渠道管理", NameEn: "Channel Management", Path: "/agents/channels"},
		{Code: "commission-plans", NameZh: "佣金方案", NameEn: "Commission Plans", Path: "/agents/commission-plans"},
		{Code: "commission-records", NameZh: "佣金记录", NameEn: "Commission Records", Path: "/agents/commission-records"},
		{Code: "agent-settlements", NameZh: "代理结算", NameEn: "Agent Settlements", Path: "/agents/settlements"},
		{Code: "agent-wallets", NameZh: "代理钱包", NameEn: "Agent Wallets", Path: "/agents/wallets"},
		{Code: "agent-operation-logs", NameZh: "代理操作日志", NameEn: "Agent Operation Logs", Path: "/agents/operation-logs"},
	}},
	{Code: "reports", NameZh: "报表中心", NameEn: "Report Center", Sort: 800, Children: []menuCatalogItem{
		{Code: "financial-report", NameZh: "财务报表", NameEn: "Financial Report", Path: "/reports/financial"},
		{Code: "game-report", NameZh: "游戏报表", NameEn: "Game Report", Path: "/reports/games"},
		{Code: "provider-report", NameZh: "Provider报表", NameEn: "Provider Report", Path: "/reports/providers"},
		{Code: "user-report", NameZh: "用户报表", NameEn: "User Report", Path: "/reports/users"},
		{Code: "channel-report", NameZh: "渠道报表", NameEn: "Channel Report", Path: "/reports/channels"},
		{Code: "agent-report", NameZh: "代理报表", NameEn: "Agent Report", Path: "/reports/agents"},
		{Code: "promotion-report", NameZh: "活动报表", NameEn: "Promotion Report", Path: "/reports/promotions"},
		{Code: "export-tasks", NameZh: "导出任务", NameEn: "Export Tasks", Path: "/reports/exports"},
	}},
	{Code: "support-center", NameZh: "支持中心", NameEn: "Support Center", Sort: 850, Children: []menuCatalogItem{
		{Code: "support-player-search", NameZh: "玩家查询", NameEn: "Player Search", Path: "/support/players"},
		{Code: "support-player-timeline", NameZh: "玩家时间线", NameEn: "Player Timeline", Path: "/support/timeline"},
		{Code: "support-order-search", NameZh: "订单查询", NameEn: "Order Search", Path: "/support/orders"},
		{Code: "support-payment-diagnostics", NameZh: "支付诊断", NameEn: "Payment Diagnostics", Path: "/support/payment-diagnostics"},
		{Code: "support-game-diagnostics", NameZh: "游戏诊断", NameEn: "Game Diagnostics", Path: "/support/game-diagnostics"},
		{Code: "support-cases", NameZh: "支持工单", NameEn: "Support Cases", Path: "/support/cases"},
		{Code: "support-exports", NameZh: "导出记录", NameEn: "Export Records", Path: "/support/exports"},
	}},
	{Code: "ops-stability", NameZh: "运维中心", NameEn: "Operations", Sort: 875, Children: []menuCatalogItem{
		{Code: "payment-monitor", NameZh: "支付监控", NameEn: "Payment Monitor", Path: "/ops/payments"},
		{Code: "provider-monitor", NameZh: "Provider监控", NameEn: "Provider Monitor", Path: "/ops/providers"},
		{Code: "maintenance-center", NameZh: "维护管理", NameEn: "Maintenance", Path: "/ops/maintenance"},
		{Code: "release-center", NameZh: "发布中心", NameEn: "Release Center", Path: "/ops/releases"},
		{Code: "system-health", NameZh: "系统健康", NameEn: "System Health", Path: "/ops/health"},
		{Code: "jobs-retries", NameZh: "任务与重试", NameEn: "Jobs & Retries", Path: "/ops/jobs"},
		{Code: "alert-center", NameZh: "告警中心", NameEn: "Alert Center", Path: "/ops/alerts"},
		{Code: "backup-center", NameZh: "备份中心", NameEn: "Backup Center", Path: "/ops/backups"},
	}},
	{Code: "system", NameZh: "系统管理", NameEn: "System Management", Sort: 900, Children: []menuCatalogItem{
		{Code: "tenants", NameZh: "租户管理", NameEn: "Tenant Management", Path: "/system/tenants"},
		{Code: "create-brand", NameZh: "创建品牌", NameEn: "Create Brand", Path: "/platform/tenants/create"},
		{Code: "brand-deliveries", NameZh: "品牌交付记录", NameEn: "Brand Deliveries", Path: "/platform/tenants/deliveries"},
		{Code: "sites", NameZh: "站点管理", NameEn: "Site Management", Path: "/system/sites"},
		{Code: "tenant-domains", NameZh: "域名管理", NameEn: "Domain Management", Path: "/platform/tenants/domains"},
		{Code: "tenant-packages", NameZh: "租户套餐", NameEn: "Tenant Packages", Path: "/platform/tenants/packages"},
		{Code: "tenant-audit", NameZh: "租户审计", NameEn: "Tenant Audit", Path: "/platform/tenants/audit"},
		{Code: "brand-config", NameZh: "品牌配置", NameEn: "Brand Config", Path: "/system/brand"},
		{Code: "system-admins", NameZh: "管理员管理", NameEn: "Administrator Management", Path: "/system/admins"},
		{Code: "system-roles", NameZh: "角色管理", NameEn: "Role Management", Path: "/system/roles"},
		{Code: "system-permissions", NameZh: "权限管理", NameEn: "Permission Management", Path: "/system/permissions"},
		{Code: "system-departments", NameZh: "部门管理", NameEn: "Department Management", Path: "/system/departments"},
		{Code: "system-menus", NameZh: "菜单管理", NameEn: "Menu Management", Path: "/system/menus"},
		{Code: "system-configs", NameZh: "系统配置", NameEn: "System Configuration", Path: "/system/config"},
		{Code: "system-countries", NameZh: "国家配置", NameEn: "Country Configuration", Path: "/system/countries"},
		{Code: "system-audit-logs", NameZh: "审计日志", NameEn: "Audit Logs", Path: "/system/audit-logs"},
		{Code: "system-login-logs", NameZh: "登录日志", NameEn: "Login Logs", Path: "/system/login-logs"},
		{Code: "system-tasks", NameZh: "定时任务", NameEn: "Scheduled Tasks", Path: "/system/tasks"},
		{Code: "system-monitoring", NameZh: "系统监控", NameEn: "System Monitoring", Path: "/system/monitor"},
	}},
}

// These names are Element Plus icon component names. They seed only blank
// records; administrators can replace them later from Menu Management.
var defaultAdminMenuIcons = map[string]string{
	"dashboard": "DataBoard", "users": "User", "user-list": "UserFilled",
	"login-records": "Clock", "device-management": "Monitor", "risk-users": "Warning",
	"funds": "Money", "wallet-accounts": "Wallet", "wallet-ledgers": "Tickets",
	"manual-adjustments": "EditPen", "deposit-orders": "CreditCard", "withdrawal-orders": "Download",
	"financial-reconciliation": "DocumentChecked", "deposit-methods": "CreditCard",
	"deposit-channels": "Connection", "deposit-amount-options": "Sort",
	"games": "Platform", "game-sync": "Refresh", "game-list": "Grid",
	"game-categories": "Collection", "lobby-sections": "Menu", "game-launch-records": "VideoPlay",
	"bets": "List", "bet-records": "DataLine",
	"rollback-records": "RefreshLeft", "abnormal-bets": "WarnTriangleFilled",
	"provider-callback-logs": "Postcard", "provider-reconciliation": "DocumentChecked", "operations": "Message", "asset-library": "FolderOpened", "banners": "Picture",
	"announcements": "Notification", "activities": "Present", "vip": "Ticket",
	"push-messages": "ChatLineSquare", "popups": "Files", "music": "Headset",
	"social-links": "Share", "h5-navigation": "Menu", "vouchers": "Tickets",
	"risk": "DocumentChecked", "risk-dashboard": "DataBoard", "risk-rules": "Lock",
	"risk-events": "WarningFilled", "review-queue": "Checked", "user-associations": "Link",
	"blacklist": "CircleCloseFilled", "withdrawal-risk": "Download", "risk-operation-logs": "Operation",
	"agents": "Share", "agent-list": "Suitcase", "agent-hierarchy": "Connection",
	"channel-management": "Position", "commission-plans": "GoldMedal", "commission-records": "Tickets",
	"agent-settlements": "Finished", "agent-wallets": "Money", "agent-operation-logs": "Operation",
	"reports": "Sort", "financial-report": "TrendCharts", "game-report": "Histogram",
	"provider-report": "PieChart", "user-report": "DataAnalysis", "channel-report": "Files",
	"agent-report": "Share", "promotion-report": "Promotion", "export-tasks": "Download",
	"support-center": "Headset", "support-player-search": "Search", "support-player-timeline": "Clock",
	"support-order-search": "Tickets", "support-payment-diagnostics": "CreditCard", "support-game-diagnostics": "Monitor",
	"support-cases": "ChatLineSquare", "support-exports": "Download", "ops-stability": "Monitor",
	"payment-monitor": "CreditCard", "provider-monitor": "Odometer", "maintenance-center": "Warning",
	"release-center": "UploadFilled", "system-health": "FirstAidKit", "jobs-retries": "Timer",
	"alert-center": "Bell", "backup-center": "FolderOpened",
	"system": "Setting", "tenants": "OfficeBuilding", "brand-deliveries": "Tickets", "tenant-domains": "Link", "tenant-packages": "Box", "tenant-audit": "DocumentChecked", "sites": "Platform", "brand-config": "Brush", "system-admins": "UserFilled", "system-roles": "Avatar",
	"system-permissions": "Key", "system-departments": "OfficeBuilding", "system-menus": "Menu",
	"system-configs": "SetUp", "system-countries": "Position", "system-audit-logs": "DocumentChecked",
	"system-login-logs": "Clock", "system-tasks": "Timer", "system-monitoring": "Odometer",
}

type NavigationRow struct {
	ID        uint64  `json:"id"`
	ParentID  *uint64 `json:"parentID" gorm:"column:parent_id"`
	Code      string  `json:"code"`
	RoutePath string  `json:"routePath" gorm:"column:route_path"`
	Icon      string  `json:"icon"`
}

func (s *Service) EnsureMenuCatalog() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		needsSync, err := menuCatalogNeedsSync(tx)
		if err != nil {
			return err
		}
		if !needsSync {
			return nil
		}
		current := time.Now()
		if err := tx.Table("game_admin_menus").Where("code=? OR route_path=?", "system-dictionaries", "/system/dictionaries").Updates(map[string]any{"deleted_at": current, "visible": 0, "enabled": 0, "updated_at": current}).Error; err != nil {
			return err
		}
		obsoleteCodes := []string{"round-records", "bonus", "register-reward", "promo-codes"}
		obsoletePaths := []string{"/bets/rounds", "/operations/bonus", "/operations/register-reward", "/operations/promo-codes"}
		if err := tx.Table("game_admin_menus").
			Where("code IN ? OR route_path IN ?", obsoleteCodes, obsoletePaths).
			Updates(map[string]any{"deleted_at": current, "visible": 0, "enabled": 0, "updated_at": current}).Error; err != nil {
			return err
		}
		for _, group := range adminMenuCatalog {
			menuType := "DIRECTORY"
			if group.Path != "" {
				menuType = "MENU"
			}
			groupRow := map[string]any{"parent_id": nil, "menu_type": menuType, "code": group.Code, "name_zh": group.NameZh, "name_en": group.NameEn, "route_path": group.Path, "sort_order": group.Sort, "visible": 1, "enabled": 1, "created_at": current, "updated_at": current}
			if err := tx.Table("game_admin_menus").Clauses(clause.Insert{Modifier: "IGNORE"}).Create(groupRow).Error; err != nil {
				return err
			}
			var parentID uint64
			if err := tx.Table("game_admin_menus").Select("id").Where("code=? AND deleted_at IS NULL", group.Code).Scan(&parentID).Error; err != nil {
				return err
			}
			for index, item := range group.Children {
				row := map[string]any{"parent_id": parentID, "menu_type": "MENU", "code": item.Code, "name_zh": item.NameZh, "name_en": item.NameEn, "route_path": item.Path, "sort_order": (index + 1) * 10, "visible": 1, "enabled": 1, "created_at": current, "updated_at": current}
				if err := tx.Table("game_admin_menus").Clauses(clause.Insert{Modifier: "IGNORE"}).Create(row).Error; err != nil {
					return err
				}
			}
		}
		if err := ensureDefaultMenuIcons(tx, current); err != nil {
			return err
		}
		return tx.Exec(`
			INSERT INTO game_admin_role_menus (role_id, menu_id, created_at, updated_at, deleted_at)
			SELECT r.id, m.id, NOW(3), NOW(3), NULL
			FROM game_admin_roles r CROSS JOIN game_admin_menus m
			WHERE r.code='super-admin' AND r.deleted_at IS NULL AND m.deleted_at IS NULL
			ON DUPLICATE KEY UPDATE deleted_at=NULL, updated_at=VALUES(updated_at)`).Error
	})
}

func menuCatalogNeedsSync(tx *gorm.DB) (bool, error) {
	expectedCodes := adminMenuCodes()
	var activeMenus int64
	if err := tx.Table("game_admin_menus").
		Where("code IN ? AND deleted_at IS NULL", expectedCodes).
		Distinct("code").
		Count(&activeMenus).Error; err != nil {
		return false, err
	}
	if activeMenus != int64(len(expectedCodes)) {
		return true, nil
	}

	var missingIcons int64
	if err := tx.Table("game_admin_menus").
		Where("code IN ? AND deleted_at IS NULL AND (icon IS NULL OR icon='') AND updated_at=created_at", menuIconCodes()).
		Count(&missingIcons).Error; err != nil {
		return false, err
	}
	if missingIcons > 0 {
		return true, nil
	}

	obsoleteCodes := []string{"system-dictionaries", "round-records", "bonus", "register-reward", "promo-codes"}
	obsoletePaths := []string{"/system/dictionaries", "/bets/rounds", "/operations/bonus", "/operations/register-reward", "/operations/promo-codes"}
	var obsoleteActive int64
	if err := tx.Table("game_admin_menus").
		Where("(code IN ? OR route_path IN ?) AND deleted_at IS NULL AND (visible=1 OR enabled=1)", obsoleteCodes, obsoletePaths).
		Count(&obsoleteActive).Error; err != nil {
		return false, err
	}
	if obsoleteActive > 0 {
		return true, nil
	}

	var missingSuperAdminMenus int64
	if err := tx.Table("game_admin_roles r").
		Joins("JOIN game_admin_menus m ON m.deleted_at IS NULL").
		Joins("LEFT JOIN game_admin_role_menus rm ON rm.role_id=r.id AND rm.menu_id=m.id AND rm.deleted_at IS NULL").
		Where("r.code=? AND r.deleted_at IS NULL AND rm.id IS NULL", "super-admin").
		Count(&missingSuperAdminMenus).Error; err != nil {
		return false, err
	}
	return missingSuperAdminMenus > 0, nil
}

func ensureDefaultMenuIcons(tx *gorm.DB, current time.Time) error {
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	if err := tx.Table("game_admin_menus").
		Select("code").
		Where("code IN ? AND deleted_at IS NULL AND (icon IS NULL OR icon='') AND updated_at=created_at", menuIconCodes()).
		Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	codes := make([]string, 0, len(rows))
	args := make([]any, 0, len(rows)*2)
	var builder strings.Builder
	builder.WriteString("CASE code")
	for _, row := range rows {
		icon := defaultAdminMenuIcons[row.Code]
		if icon == "" {
			continue
		}
		codes = append(codes, row.Code)
		builder.WriteString(" WHEN ? THEN ?")
		args = append(args, row.Code, icon)
	}
	if len(codes) == 0 {
		return nil
	}
	builder.WriteString(" ELSE icon END")
	return tx.Table("game_admin_menus").
		Where("code IN ? AND deleted_at IS NULL AND (icon IS NULL OR icon='') AND updated_at=created_at", codes).
		Updates(map[string]any{"icon": gorm.Expr(builder.String(), args...), "updated_at": current}).Error
}

func adminMenuCodes() []string {
	codes := make([]string, 0, len(adminMenuCatalog)*4)
	for _, group := range adminMenuCatalog {
		codes = append(codes, group.Code)
		for _, child := range group.Children {
			codes = append(codes, child.Code)
		}
	}
	return codes
}

func menuIconCodes() []string {
	codes := make([]string, 0, len(defaultAdminMenuIcons))
	for code := range defaultAdminMenuIcons {
		codes = append(codes, code)
	}
	return codes
}

var platformOnlyMenuCodes = map[string]bool{
	"tenants": true, "create-brand": true, "brand-deliveries": true, "tenant-domains": true,
	"tenant-packages": true, "tenant-audit": true, "game-sync": true, "provider-callback-logs": true,
	"provider-reconciliation": true, "provider-report": true, "system-admins": true,
	"system-roles": true, "system-permissions": true, "system-departments": true,
	"system-menus": true, "system-configs": true, "system-countries": true,
	"system-tasks": true, "system-monitoring": true,
	"system-health": true, "jobs-retries": true, "alert-center": true, "backup-center": true,
}

var tenantOnlyMenuCodes = map[string]bool{
	"user-list": true, "login-records": true, "device-management": true, "risk-users": true,
	"wallet-accounts": true, "wallet-ledgers": true, "manual-adjustments": true,
	"deposit-orders": true, "withdrawal-orders": true, "financial-reconciliation": true,
	"deposit-methods": true, "deposit-channels": true, "deposit-amount-options": true,
	"lobby-sections": true, "game-launch-records": true, "bet-records": true,
	"rollback-records": true, "abnormal-bets": true, "banners": true, "announcements": true,
	"activities": true, "vip": true, "push-messages": true, "popups": true, "music": true,
	"social-links": true, "h5-navigation": true, "vouchers": true, "risk-dashboard": true,
	"risk-rules": true, "risk-events": true, "review-queue": true, "user-associations": true,
	"blacklist": true, "withdrawal-risk": true, "risk-operation-logs": true, "agent-list": true,
	"agent-hierarchy": true, "channel-management": true, "commission-plans": true,
	"commission-records": true, "agent-settlements": true, "agent-wallets": true,
	"agent-operation-logs": true, "financial-report": true, "game-report": true,
	"user-report": true, "channel-report": true, "agent-report": true, "promotion-report": true,
	"brand-config": true,
}

func menuAllowedForPortal(code string, platform bool) bool {
	if platform {
		return !tenantOnlyMenuCodes[code]
	}
	return !platformOnlyMenuCodes[code]
}

func filterNavigationPortal(rows []NavigationRow, platform bool) []NavigationRow {
	allowedIDs := map[uint64]bool{}
	for _, row := range rows {
		if row.RoutePath != "" && menuAllowedForPortal(row.Code, platform) {
			allowedIDs[row.ID] = true
			if row.ParentID != nil {
				allowedIDs[*row.ParentID] = true
			}
		}
	}
	filtered := make([]NavigationRow, 0, len(rows))
	for _, row := range rows {
		if allowedIDs[row.ID] {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func (s *Service) Navigation(actor Actor, superAdmin bool) (map[string]any, error) {
	rows := []NavigationRow{}
	if err := s.db.Table("game_admin_menus").Select("id,parent_id,code,route_path,icon").Where("enabled=1 AND visible=1 AND deleted_at IS NULL").Order("sort_order,id").Find(&rows).Error; err != nil {
		return nil, err
	}
	rows = filterNavigationPortal(rows, actor.Platform)
	if superAdmin {
		return map[string]any{"list": rows}, nil
	}
	assigned := []uint64{}
	if err := s.db.Table("game_admin_account_roles ar").Distinct("rm.menu_id").
		Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.status=1 AND r.deleted_at IS NULL").
		Joins("JOIN game_admin_role_menus rm ON rm.role_id=r.id AND rm.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", actor.ID).Pluck("rm.menu_id", &assigned).Error; err != nil {
		return nil, err
	}
	allowed := make(map[uint64]bool, len(assigned)*2)
	for _, id := range assigned {
		allowed[id] = true
	}
	for _, row := range rows {
		if allowed[row.ID] && row.ParentID != nil {
			allowed[*row.ParentID] = true
		}
	}
	filtered := make([]NavigationRow, 0, len(rows))
	for _, row := range rows {
		if allowed[row.ID] {
			filtered = append(filtered, row)
		}
	}
	return map[string]any{"list": filtered}, nil
}
