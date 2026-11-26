import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import i18n from './i18n'
import router from './router'
import './styles/index.css'

createApp(App).use(router).use(ElementPlus).use(i18n).mount('#app')
