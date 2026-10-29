<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import {
  Activity,
  ChevronDown,
  Container,
  Cpu,
  Gauge,
  HardDrive,
  Plus,
  Router,
  Settings,
  Shield,
  type LucideIcon,
} from 'lucide-vue-next'
import { api } from '@/api/client'
import { core, type CoreEndpoint } from '@/api/client'
import { cn } from '@/lib/utils'

const route = useRoute()

interface NavGroup {
  key: string
  label: string
  icon: LucideIcon
  to?: string
  children?: { to: string; label: string; icon?: LucideIcon }[]
}

const navGroups: NavGroup[] = [
  { key: 'dashboard', label: 'Dashboard', icon: Gauge, to: '/' },
  {
    key: 'hardware',
    label: 'Hardware',
    icon: Cpu,
    children: [
      { to: '/cpu', label: 'CPU', icon: Cpu },
      { to: '/memory', label: 'Memory' },
      { to: '/bandwidth', label: 'Network', icon: Router },
      { to: '/disk', label: 'Disks', icon: HardDrive },
    ],
  },
  {
    key: 'agent',
    label: 'Agent',
    icon: Router,
    children: [
      { to: '/devices', label: 'Devices/machines' },
      { to: '/devices/scan', label: 'Add', icon: Plus },
    ],
  },
  {
    key: 'system',
    label: 'System',
    icon: Settings,
    children: [
      { to: '/temperature', label: 'Temperature' },
      { to: '/motherboard', label: 'Motherboard' },
      { to: '/system', label: 'System info' },
    ],
  },
  {
    key: 'activity',
    label: 'Activity',
    icon: Activity,
    children: [
      { to: '/apps', label: 'Apps' },
      { to: '/events', label: 'Events' },
      { to: '/actions', label: 'Actions' },
    ],
  },
]

function isGroupActive(group: NavGroup) {
  const paths = group.to ? [group.to] : (group.children ?? []).map((c) => c.to)
  return paths.some((p) => (p === '/' ? route.path === '/' : route.path === p || route.path.startsWith(`${p}/`)))
}

const openGroups = ref<Record<string, boolean>>({})

function toggleGroup(key: string) {
  openGroups.value[key] = !openGroups.value[key]
}

const apiOk = ref<boolean | null>(null)
const machines = ref<CoreEndpoint[]>([])
const machinesError = ref('')

watch(() => route.path, () => {
  for (const group of navGroups) {
    if (isGroupActive(group)) openGroups.value[group.key] = true
  }
}, { immediate: true })

function navLinkClass(active: boolean, size: 'sm' | 'md' = 'md') {
  return cn(
    'flex items-center gap-2 rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
    size === 'md' ? 'px-3 py-2 text-sm' : 'px-2 py-1.5 text-xs',
    active && 'bg-muted font-medium text-foreground',
  )
}

let healthTimer: ReturnType<typeof setInterval> | undefined
let machinesTimer: ReturnType<typeof setInterval> | undefined

async function loadMachines() {
  try {
    const res = await core.endpoints()
    machines.value = res.endpoints
    machinesError.value = res.agentError ?? ''
  } catch (e) {
    machinesError.value = e instanceof Error ? e.message : 'Failed to load machines'
  }
}

async function checkApi() {
  try {
    await api.dashboard()
    apiOk.value = true
  } catch {
    apiOk.value = false
  }
}

onMounted(() => {
  void checkApi()
  healthTimer = setInterval(checkApi, 2000)
  void loadMachines()
  machinesTimer = setInterval(loadMachines, 5000)
})

onUnmounted(() => {
  if (healthTimer) clearInterval(healthTimer)
  if (machinesTimer) clearInterval(machinesTimer)
})
</script>

<template>
  <div class="flex h-screen min-h-0">
    <aside
      class="flex w-56 shrink-0 flex-col border-r border-border bg-card/70 backdrop-blur"
    >
      <div class="flex items-center gap-2 border-b border-border px-4 py-4">
        <img src="/favicon.svg" alt="" class="h-8 w-8 shrink-0" width="32" height="32" />
        <div>
          <div class="text-sm font-semibold tracking-wide">Nevarix</div>
          <div class="text-[11px] text-muted-foreground">
            {{ apiOk === false ? 'API offline' : 'Host monitor' }}
          </div>
        </div>
      </div>

      <nav class="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
        <template v-for="group in navGroups" :key="group.key">
          <RouterLink
            v-if="group.to"
            :to="group.to"
            custom
            v-slot="{ isActive, href, navigate }"
          >
            <a :href="href" :class="navLinkClass(isActive)" @click="navigate">
              <component :is="group.icon" class="h-4 w-4 shrink-0" />
              {{ group.label }}
            </a>
          </RouterLink>

          <div v-else>
            <button
              type="button"
              :class="navLinkClass(isGroupActive(group), 'md') + ' w-full'"
              @click="toggleGroup(group.key)"
            >
              <component :is="group.icon" class="h-4 w-4 shrink-0" />
              <span class="flex-1 text-left">{{ group.label }}</span>
              <ChevronDown
                class="h-3.5 w-3.5 shrink-0 transition-transform"
                :class="openGroups[group.key] && 'rotate-180'"
              />
            </button>
            <div
              v-if="openGroups[group.key]"
              class="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-border pl-2"
            >
              <RouterLink
                v-for="child in group.children"
                :key="child.to"
                :to="child.to"
                custom
                v-slot="{ isActive, href, navigate }"
              >
                <a
                  :href="href"
                  :class="navLinkClass(isActive, 'sm')"
                  @click="navigate"
                >
                  <component
                    :is="child.icon"
                    v-if="child.icon"
                    class="h-3.5 w-3.5 shrink-0"
                  />
                  {{ child.label }}
                </a>
              </RouterLink>
            </div>
          </div>
        </template>

        <RouterLink to="/docker" custom v-slot="{ isActive, href, navigate }">
          <a :href="href" :class="navLinkClass(isActive)" @click="navigate">
            <Container class="h-4 w-4 shrink-0" />
            Docker
          </a>
        </RouterLink>

        <RouterLink to="/firewall" custom v-slot="{ isActive, href, navigate }">
          <a :href="href" :class="navLinkClass(isActive)" @click="navigate">
            <Shield class="h-4 w-4 shrink-0" />
            Firewall
          </a>
        </RouterLink>

        <RouterLink to="/settings" custom v-slot="{ isActive, href, navigate }">
          <a :href="href" :class="navLinkClass(isActive)" @click="navigate">
            <Settings class="h-4 w-4 shrink-0" />
            Settings
          </a>
        </RouterLink>
      </nav>

      <div class="border-t border-border p-3 text-[11px] text-muted-foreground">
        <div class="mb-1 flex items-center gap-1.5">
          <span
            class="inline-block h-1.5 w-1.5 rounded-full"
            :class="apiOk ? 'bg-emerald-500' : apiOk === false ? 'bg-destructive' : 'bg-muted'"
          />
          {{ apiOk ? 'nevarix-api connected' : apiOk === false ? 'waiting for nevarix-api' : 'checking api…' }}
        </div>
        Created by Armin Dashti
      </div>
    </aside>

    <aside class="flex w-60 shrink-0 flex-col border-r border-border bg-card/50">
      <div class="border-b border-border px-4 py-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Machines
      </div>
      <div class="flex-1 overflow-y-auto p-2">
        <div v-if="machinesError" class="px-2 py-1 text-xs text-destructive">{{ machinesError }}</div>
        <div v-else-if="machines.length === 0" class="px-2 py-1 text-xs text-muted-foreground">
          No machines reported by agents
        </div>
        <RouterLink
          v-for="m in machines"
          :key="m.hostname + (m.ip ?? '')"
          to="/endpoints"
          custom
          v-slot="{ href, navigate, isActive }"
        >
          <a
            :href="href"
            :class="navLinkClass(isActive, 'sm')"
            class="block"
            @click="navigate"
          >
            <div class="truncate font-medium">{{ m.hostname }}</div>
            <div class="truncate text-[11px] text-muted-foreground">{{ m.ip ?? 'unknown ip' }}</div>
          </a>
        </RouterLink>
      </div>
    </aside>

    <main class="min-w-0 flex-1 overflow-auto p-5">
      <RouterView />
    </main>
  </div>
</template>
