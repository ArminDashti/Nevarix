<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import {
  ChevronDown,
  Container,
  Cpu,
  Gauge,
  Network,
  Settings,
  Shield,
} from 'lucide-vue-next'
import { api } from '@/api/client'
import { cn } from '@/lib/utils'

const route = useRoute()

const systemChildren = [
  { to: '/cpu', label: 'CPU' },
  { to: '/memory', label: 'Memory' },
  { to: '/disk', label: 'Disk' },
  { to: '/temperature', label: 'Temperature' },
  { to: '/motherboard', label: 'Motherboard' },
  { to: '/system', label: 'Overview' },
]

const networkChildren = [
  { to: '/bandwidth', label: 'Bandwidth' },
  { to: '/endpoints', label: 'Endpoints' },
  { to: '/endpoints/scan', label: 'Scan' },
]

const onSystemRoute = computed(() =>
  systemChildren.some((c) => route.path === c.to || route.path.startsWith(`${c.to}/`)),
)
const onNetworkRoute = computed(() =>
  networkChildren.some((c) => route.path === c.to || route.path.startsWith(`${c.to}/`)),
)

const systemOpen = ref(onSystemRoute.value)
const networkOpen = ref(onNetworkRoute.value)
const apiOk = ref<boolean | null>(null)

watch(onSystemRoute, (v) => {
  if (v) systemOpen.value = true
})
watch(onNetworkRoute, (v) => {
  if (v) networkOpen.value = true
})

function navLinkClass(active: boolean, size: 'sm' | 'md' = 'md') {
  return cn(
    'flex items-center gap-2 rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
    size === 'md' ? 'px-3 py-2 text-sm' : 'px-2 py-1.5 text-xs',
    active && 'bg-muted font-medium text-foreground',
  )
}

let healthTimer: ReturnType<typeof setInterval> | undefined

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
})

onUnmounted(() => {
  if (healthTimer) clearInterval(healthTimer)
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
        <RouterLink to="/" custom v-slot="{ isExactActive, href, navigate }">
          <a
            :href="href"
            :class="navLinkClass(isExactActive)"
            @click="navigate"
          >
            <Gauge class="h-4 w-4 shrink-0" />
            Dashboard
          </a>
        </RouterLink>

        <div>
          <button
            type="button"
            :class="navLinkClass(onSystemRoute, 'md') + ' w-full'"
            @click="systemOpen = !systemOpen"
          >
            <Cpu class="h-4 w-4 shrink-0" />
            <span class="flex-1 text-left">System</span>
            <ChevronDown
              class="h-3.5 w-3.5 shrink-0 transition-transform"
              :class="systemOpen && 'rotate-180'"
            />
          </button>
          <div
            v-if="systemOpen"
            class="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-border pl-2"
          >
            <RouterLink
              v-for="child in systemChildren"
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
                {{ child.label }}
              </a>
            </RouterLink>
          </div>
        </div>

        <div>
          <button
            type="button"
            :class="navLinkClass(onNetworkRoute, 'md') + ' w-full'"
            @click="networkOpen = !networkOpen"
          >
            <Network class="h-4 w-4 shrink-0" />
            <span class="flex-1 text-left">Network</span>
            <ChevronDown
              class="h-3.5 w-3.5 shrink-0 transition-transform"
              :class="networkOpen && 'rotate-180'"
            />
          </button>
          <div
            v-if="networkOpen"
            class="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-border pl-2"
          >
            <RouterLink
              v-for="child in networkChildren"
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
                {{ child.label }}
              </a>
            </RouterLink>
          </div>
        </div>

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

    <main class="min-w-0 flex-1 overflow-auto p-5">
      <RouterView />
    </main>
  </div>
</template>
