import { createRouter, createWebHistory } from 'vue-router'
import UploadView from '@/views/UploadView.vue'
import DashboardView from '@/views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: UploadView },
    { path: '/dashboard', component: DashboardView },
  ],
})

export default router

