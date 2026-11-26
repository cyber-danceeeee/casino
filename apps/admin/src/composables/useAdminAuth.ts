import { computed, ref } from "vue";
import {
  fetchCurrentAdmin,
  logoutAdmin,
  type AdminProfile,
} from "@/api/adminAuth";
import {
  clearAdminActor,
  setAdminActor,
} from "@/composables/useAdminPermissions";

function readStoredProfile(): AdminProfile | null {
  const stored = window.localStorage.getItem("casino-admin-profile");
  if (!stored) return null;
  try {
    return JSON.parse(stored) as AdminProfile;
  } catch {
    window.localStorage.removeItem("casino-admin-profile");
    return null;
  }
}

const profile = ref<AdminProfile | null>(readStoredProfile());

export function useAdminAuth() {
  const authenticated = computed(() =>
    Boolean(window.localStorage.getItem("casino-admin-token")),
  );

  async function establishSession(token: string, value: AdminProfile) {
    window.localStorage.setItem("casino-admin-token", token);
    window.localStorage.setItem("casino-admin-profile", JSON.stringify(value));
    profile.value = value;
    const actor = await fetchCurrentAdmin();
    setAdminActor(actor);
  }

  async function refreshSession() {
    if (!window.localStorage.getItem("casino-admin-token")) return null;
    const actor = await fetchCurrentAdmin();
    setAdminActor(actor);
    if (!profile.value) {
      profile.value = {
        id: actor.id,
        username: actor.username,
        displayName: actor.displayName || actor.username,
        tenantID: actor.tenantID,
        siteID: actor.siteID,
        defaultSiteID: actor.defaultSiteID,
        adminType: actor.adminType,
        mustChangePassword: actor.mustChangePassword,
      };
    }
    return actor;
  }

  async function logout() {
    try {
      await logoutAdmin();
    } finally {
      window.localStorage.removeItem("casino-admin-token");
      window.localStorage.removeItem("casino-admin-profile");
      profile.value = null;
      clearAdminActor();
    }
  }

  return {
    profile: computed(() => profile.value),
    authenticated,
    establishSession,
    refreshSession,
    logout,
  };
}
