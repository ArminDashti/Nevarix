import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import EndpointsView from '@/views/EndpointsView.vue'
import ScanView from '@/views/ScanView.vue'
import SettingsView from '@/views/SettingsView.vue'
import AgentsView from '@/views/AgentsView.vue'
import DevicesView from '@/views/DevicesView.vue'
import DeviceDetailView from '@/views/DeviceDetailView.vue'
import DeviceScanView from '@/views/DeviceScanView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    { path: '/endpoints', name: 'endpoints', component: EndpointsView },
    { path: '/endpoints/scan', name: 'endpoints-scan', component: ScanView },
    { path: '/agents', name: 'agents', component: AgentsView },
    { path: '/agents/:id', name: 'agent', component: AgentsView },
    { path: '/devices', name: 'devices', component: DevicesView },
    { path: '/devices/scan', name: 'devices-scan', component: DeviceScanView },
    { path: '/devices/:deviceId', name: 'device-detail', component: DeviceDetailView },
    { path: '/settings', name: 'settings', component: SettingsView },
  ],
})

export default router
