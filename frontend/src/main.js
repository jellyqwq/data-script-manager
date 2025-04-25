// src/main.js
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import '@vue-js-cron/element-plus/dist/element-plus.css'

// registers the component globally
// registered name: CronElementPlus
import CronElementPlusPlugin from '@vue-js-cron/element-plus'


const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.use(CronElementPlusPlugin)
app.mount('#app')

