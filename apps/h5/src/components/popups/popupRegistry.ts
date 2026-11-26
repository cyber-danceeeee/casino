import { defineAsyncComponent, type Component } from 'vue'

const popupRegistry: Record<string, Component> = {
  DEPOSIT_POPUP: defineAsyncComponent(() => import('./DepositPopup.vue')),
  WELCOME_BONUS_POPUP: defineAsyncComponent(() => import('./WelcomeBonusPopup.vue')),
  CASHBACK_POPUP: defineAsyncComponent(() => import('./CashbackPopup.vue')),
  VIP_UPGRADE_POPUP: defineAsyncComponent(() => import('./VipUpgradePopup.vue')),
  APP_DOWNLOAD_POPUP: defineAsyncComponent(() => import('./AppDownloadPopup.vue')),
}

export const resolvePopupComponent = (componentCode: string) => popupRegistry[componentCode] || null
export const supportedPopupComponentCodes = Object.freeze(Object.keys(popupRegistry))
