import { createRouter, createWebHistory } from 'vue-router'
import WebPages from '../views/WebPages.vue'
import MonitorConfig from '../views/MonitorConfig.vue'
import SubscribeConfig from '../views/SubscribeConfig.vue'
import Announcements from '../views/Announcements.vue'

const routes = [
  { path: '/', redirect: '/web-pages' },
  { path: '/web-pages', component: WebPages },
  { path: '/monitor-config', component: MonitorConfig },
  { path: '/subscribe-config', component: SubscribeConfig },
  { path: '/announcements', component: Announcements }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
