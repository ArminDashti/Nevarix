import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import CpuView from '@/views/CpuView.vue'
import MemoryView from '@/views/MemoryView.vue'
import DiskView from '@/views/DiskView.vue'
import BandwidthView from '@/views/BandwidthView.vue'
import DockerView from '@/views/DockerView.vue'
import EndpointsView from '@/views/EndpointsView.vue'
import ScanView from '@/views/ScanView.vue'
import SettingsView from '@/views/SettingsView.vue'
import SystemView from '@/views/SystemView.vue'
import MotherboardView from '@/views/MotherboardView.vue'
import TemperatureView from '@/views/TemperatureView.vue'
import FirewallView from '@/views/FirewallView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    { path: '/cpu', name: 'cpu', component: CpuView },
    { path: '/memory', name: 'memory', component: MemoryView },
    { path: '/disk', name: 'disk', component: DiskView },
    { path: '/bandwidth', name: 'bandwidth', component: BandwidthView },
    { path: '/docker', name: 'docker', component: DockerView },
    { path: '/endpoints', name: 'endpoints', component: EndpointsView },
    { path: '/endpoints/scan', name: 'endpoints-scan', component: ScanView },
    { path: '/system', name: 'system', component: SystemView },
    { path: '/motherboard', name: 'motherboard', component: MotherboardView },
    { path: '/temperature', name: 'temperature', component: TemperatureView },
    { path: '/firewall', name: 'firewall', component: FirewallView },
    { path: '/settings', name: 'settings', component: SettingsView },
  ],
})

export default router
