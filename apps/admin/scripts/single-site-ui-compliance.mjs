import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const read = (path) => readFileSync(resolve(root, path), "utf8");

const menu = read("src/data/adminMenu.ts");
for (const key of ["site-support","users","funds","games","operations","agents","risk","reports","brand","site-ops"]) {
 assert.match(menu,new RegExp(`key: ["']${key}["']`),`missing administration menu: ${key}`);
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
assert.match(maintenance, /v-if=["']isPlatform["'][^>]*fields\.siteID/, "H5 maintenance site scope must be implicit");

const monitor = read("src/views/operations/OperationsMonitorPage.vue");
assert.match(monitor, /paymentDetailByNumber/, "payment diagnostics must search by business order number");
assert.doesNotMatch(monitor, /Number\(diagnostic\.order/, "payment diagnostics must not parse a database ID");

const routes = read("src/router/index.ts");
assert.match(routes, /players\/:playerNo/, "business player route is missing");

assert.doesNotMatch(read("src/layouts/AdminLayout.vue"), /BrandDraftStatusBar/, "obsolete publication UI remains");
assert.doesNotMatch(routes, /release-center|brand-publish|brand-preview/, "obsolete publication route remains");
const dashboard = read("src/views/brand/BrandDashboard.vue");
for (const field of ["depositAmount", "withdrawalAmount", "totalBet", "ggr", "onlineUsers"]) {
  assert.match(dashboard, new RegExp(field), `dashboard missing ${field}`);
}
assert.match(dashboard, /overview(?:\.value)?\??\.currencyCode/, "dashboard must use H5 currency");
assert.match(dashboard, /overview(?:\.value)?\??\.timezone/, "dashboard must use H5 timezone");
assert.match(dashboard, /value="custom"/, "dashboard must support a custom date range");

const funds = read("src/views/funds/FundsPage.vue");
assert.doesNotMatch(funds, /tr\("(?:Adjustment|Platform Order) ID"\)/, "dangerous finance confirmations must use business numbers");

const reports = read("src/views/reports/config.ts");
assert.match(reports, /f\('playerNo'\)/, "player reports must search by player number");
assert.match(reports, /c\('gameCode'/, "game reports must display a game code");

const stability = read("src/views/operations/PlatformStabilityPage.vue");
assert.match(stability, /subTab\.value = 'retryRecords'/, "H5 jobs must default to H5-scoped retry records");

const gameCenter = read("src/views/games/GameCenterPage.vue");
assert.match(gameCenter, /const isPlatform = computed/, "game center must distinguish platform and H5 workflows");
assert.doesNotMatch(gameCenter, /usesDraftWorkflow|sites\.publish|draftConfiguration/, "game center still stages changes");
assert.match(gameCenter, /savedImmediate/, "game saves must communicate immediate effect");
const messages = read("src/i18n/messages.ts");
for (const phrase of ["玩家编号", "Player Number", "游戏编码", "Game Code"]) {
  assert.ok(messages.includes(phrase), `missing bilingual label: ${phrase}`);
}

assert.match(menu, /return singleSiteMenuGroups/, "navigation must use the single-site menu");
const login = read("src/views/auth/AdminLogin.vue");
assert.doesNotMatch(login, /H5Code|isTenantPortal|portalType|router-link/, "login still exposes a H5 or portal selector");
assert.match(routes, /name: "admin-login"/, "the unified login route is missing");
assert.doesNotMatch(routes, /path: "system\/H5s\/:id"/, "H5 detail is still routable");
const settings = read("src/views/brand/BrandSettingsPage.vue");
assert.match(settings, /systemAPI.brand.save/, "H5 settings must save to the bound site");
assert.match(settings, /form.h5Domain/, "the H5 address field is missing");
assert.doesNotMatch(settings, /v-model="form.(?:countryCode|currencyCode)" disabled/, "market fields remain locked");
console.log("Single-site administration and operational UI checks passed.");
