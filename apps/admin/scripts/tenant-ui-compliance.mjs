import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const read = (path) => readFileSync(resolve(root, path), "utf8");

const menu = read("src/data/adminMenu.ts");
for (const key of [
  "tenant-support",
  "tenant-players",
  "tenant-finance",
  "tenant-games",
  "tenant-marketing",
  "agents",
  "risk",
  "reports",
  "tenant-brand",
  "tenant-ops",
]) {
  assert.match(menu, new RegExp(`key: ["']${key}["']`), `missing tenant menu section: ${key}`);
}

const userList = read("src/views/users/UserListPage.vue");
assert.doesNotMatch(userList, /tr\(["']User ID["']\)/, "player list renders User ID");
assert.match(userList, /tr\(["']Player Number["']\)/, "player list must render Player Number");

const agentConfig = read("src/views/agents/config.ts");
for (const filter of ["agentID", "recordID", "settlementID", "walletID", "userID", "targetID"]) {
  assert.doesNotMatch(agentConfig, new RegExp(`f\\(["']${filter}["']\\)`), `agent filter exposes ${filter}`);
}

const assetLibrary = read("src/views/operations/AssetLibraryPage.vue");
assert.doesNotMatch(assetLibrary, /prop=["']recordID["']/, "asset references expose record ID");

const popup = read("src/views/operations/PopupManagementPage.vue");
assert.match(popup, /componentCode/, "popup campaign must use a stable H5 component code");
assert.match(popup, /livePreviewURL/, "popup campaign must use the real H5 preview route");
assert.doesNotMatch(popup, /Popup Template ID|Popup ID/, "popup campaign exposes internal IDs");

const banner = read("src/views/operations/BannerManagementPage.vue");
assert.match(banner, /H5PreviewFrame/, "banner management must render a real H5 preview");
assert.match(banner, /actionTargetOptions/, "banner targets must use business selectors");

const maintenance = read("src/views/operations/MaintenanceCenterPage.vue");
assert.match(maintenance, /v-if=["']isPlatform["'][^>]*fields\.siteID/, "tenant maintenance site scope must be implicit");

const monitor = read("src/views/operations/OperationsMonitorPage.vue");
assert.match(monitor, /paymentDetailByNumber/, "payment diagnostics must search by business order number");
assert.doesNotMatch(monitor, /Number\(diagnostic\.order/, "payment diagnostics must not parse a database ID");

const routes = read("src/router/index.ts");
assert.match(routes, /players\/:playerNo/, "business player route is missing");

const publishBar = read("src/components/brand/BrandDraftStatusBar.vue");
for (const field of ["unpublishedChanges", "publishedVersion", "draftVersion", "pendingChanges"]) {
  assert.match(publishBar, new RegExp(field), `publish bar missing ${field}`);
}

const dashboard = read("src/views/brand/BrandDashboard.vue");
for (const field of ["depositAmount", "withdrawalAmount", "totalBet", "ggr", "onlineUsers"]) {
  assert.match(dashboard, new RegExp(field), `dashboard missing ${field}`);
}
assert.match(dashboard, /overview(?:\.value)?\??\.currencyCode/, "dashboard must use tenant currency");
assert.match(dashboard, /overview(?:\.value)?\??\.timezone/, "dashboard must use tenant timezone");
assert.match(dashboard, /value="custom"/, "dashboard must support a custom date range");

const funds = read("src/views/funds/FundsPage.vue");
assert.doesNotMatch(funds, /tr\("(?:Adjustment|Platform Order) ID"\)/, "dangerous finance confirmations must use business numbers");

const reports = read("src/views/reports/config.ts");
assert.match(reports, /f\('playerNo'\)/, "player reports must search by player number");
assert.match(reports, /c\('gameCode'/, "game reports must display a game code");

const stability = read("src/views/operations/PlatformStabilityPage.vue");
assert.match(stability, /subTab\.value = 'retryRecords'/, "tenant jobs must default to tenant-scoped retry records");

const gameCenter = read("src/views/games/GameCenterPage.vue");
assert.match(gameCenter, /const isPlatform = computed/, "game center must distinguish platform and tenant workflows");
assert.match(gameCenter, /usesDraftWorkflow/, "tenant game changes must retain the draft workflow");
assert.match(gameCenter, /savedImmediate/, "platform game changes must communicate immediate effect");
assert.match(gameCenter, /"publishedState", "draftState", "hasChanges", "draftUpdatedAt"/, "platform game list must hide draft-only columns");

const messages = read("src/i18n/messages.ts");
for (const phrase of ["玩家编号", "Player Number", "游戏编码", "Game Code", "发布版本", "Published Version"]) {
  assert.ok(messages.includes(phrase), `missing bilingual label: ${phrase}`);
}

console.log("Tenant Admin UI compliance checks passed.");
