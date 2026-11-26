import { createApp } from 'vue'
import './tailwind.css'
import './style.scss'
import App from './App.vue'
import router from './router'
import {
  cacheApplicationShell,
  configureAssetCache,
  registerAssetCache,
} from './services/assetCache'
import { initializePWAInstall } from './services/pwa'
import { loadRuntimeConfig } from './services/runtimeConfig'
import { installTapFeedback } from './services/tapFeedback'
import { initializeNativePushRegistration } from './services/nativePushRegistration'

const bootstrap = async () => {
  initializePWAInstall()
  await registerAssetCache()
  const config = await loadRuntimeConfig()
  await configureAssetCache(config.resourceVersion)
  void cacheApplicationShell()
  installTapFeedback()
  initializeNativePushRegistration()
  createApp(App).use(router).mount('#app')
}

void bootstrap()
