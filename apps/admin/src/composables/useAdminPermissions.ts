import { computed, ref } from "vue";
import { fetchAdminActor, type AdminActor } from "@/api/adminUsers";
import { systemAPI } from "@/api/adminSystem";

const actor = ref<AdminActor | null>(null);
const navigationPaths = ref<Set<string>>(new Set());
const orderedNavigationPaths = ref<string[]>([]);
const navigationIcons = ref<Map<string, string>>(new Map());
const navigationReady = ref(false);
let request: Promise<AdminActor> | null = null;
let navigationRequest: Promise<void> | null = null;

async function loadNavigation(force = false) {
  if (navigationReady.value && !force) return;
  if (!navigationRequest) {
    navigationRequest = systemAPI.navigation()
      .then((result) => {
        const paths = (result.list || [])
          .map((item) => item.routePath)
          .filter(Boolean);
        navigationPaths.value = new Set(paths);
        orderedNavigationPaths.value = paths;
        navigationIcons.value = new Map(
          (result.list || []).flatMap((item) => {
            const entries: Array<[string, string]> = [];
            if (item.code && item.icon) entries.push([`code:${item.code}`, item.icon]);
            if (item.routePath && item.icon) entries.push([`path:${item.routePath}`, item.icon]);
            return entries;
          }),
        );
        navigationReady.value = true;
      })
      .finally(() => (navigationRequest = null));
  }
  await navigationRequest;
}

export function useAdminPermissions() {
  async function loadPermissions(forceNavigation = false) {
    if (actor.value) {
      await loadNavigation(forceNavigation);
      return actor.value;
    }
    if (!request)
      request = fetchAdminActor()
        .then((value) => (actor.value = value))
        .finally(() => (request = null));
    const value = await request;
    await loadNavigation(forceNavigation);
    return value;
  }
  const can = (permission: string) =>
    Boolean(
      actor.value?.permissions.includes("*") ||
      actor.value?.permissions.includes(permission),
    );
  const hasNavigation = (path: string) => navigationPaths.value.has(path);
  const navigationIcon = (path: string, code: string) =>
    navigationIcons.value.get(`path:${path}`) ||
    navigationIcons.value.get(`code:${code}`) ||
    "";
  const firstNavigationPath = () =>
    orderedNavigationPaths.value.find((path) => path !== "/") || "";
  const reloadNavigation = () => loadNavigation(true);
  return {
    actor: computed(() => actor.value),
    navigationReady: computed(() => navigationReady.value),
    can,
    hasNavigation,
    navigationIcon,
    firstNavigationPath,
    loadPermissions,
    reloadNavigation,
  };
}

export function setAdminActor(value: AdminActor) {
  actor.value = value;
}

export function clearAdminActor() {
  actor.value = null;
  navigationPaths.value = new Set();
  orderedNavigationPaths.value = [];
  navigationIcons.value = new Map();
  navigationReady.value = false;
  request = null;
  navigationRequest = null;
}
