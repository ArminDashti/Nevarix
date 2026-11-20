<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { Power, RotateCcw, Trash2 } from 'lucide-vue-next'
import { core, type AgentSummary } from '@/api/client'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatAgentId, loadAgentIdFormat, type AgentIdFormat } from '@/lib/agentIdFormat'

const route = useRoute()
const agents = ref<AgentSummary[]>([])
const error = ref('')
const actionError = ref('')
const idFormat = ref<AgentIdFormat>(loadAgentIdFormat())
let timer: ReturnType<typeof setInterval> | undefined

const rows = computed(() => {
  const id = typeof route.params.id === 'string' ? route.params.id : ''
  if (!id) return agents.value
  return agents.value.filter((a) => a.agentId === id)
})

function stamp(iso: string) {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

function pct(n: number) {
  return `${n.toFixed(1)}%`
}

function light(status: AgentSummary['status']) {
  if (status === 'connected') return 'bg-emerald-500'
  if (status === 'unstable') return 'bg-yellow-400'
  return 'bg-red-500'
}

async function load() {
  try {
    const res = await core.agents()
    agents.value = res.agents ?? []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load agents'
  }
}

async function run(agent: AgentSummary, action: 'restart' | 'shutdown' | 'delete') {
  actionError.value = ''
  const label = action[0].toUpperCase() + action.slice(1)
  if (!window.confirm(`${label} ${agent.hostname || agent.agentId}?`)) return
  try {
    if (action === 'delete') await core.deleteAgent(agent.agentId)
    else await core.agentAction(agent.agentId, action)
    await load()
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : 'Action failed'
  }
}

function onStorage() {
  idFormat.value = loadAgentIdFormat()
}

onMounted(() => {
  void load()
  timer = setInterval(load, 5000)
  window.addEventListener('storage', onStorage)
  window.addEventListener('nevarix-agent-id-format', onStorage)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  window.removeEventListener('storage', onStorage)
  window.removeEventListener('nevarix-agent-id-format', onStorage)
})
</script>

<template>
  <div class="space-y-3">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <p v-if="actionError" class="text-sm text-destructive">{{ actionError }}</p>
    <div class="overflow-x-auto rounded-md border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>CPU</TableHead>
            <TableHead>Mem</TableHead>
            <TableHead>Disks</TableHead>
            <TableHead>OS</TableHead>
            <TableHead>IP</TableHead>
            <TableHead>Latency</TableHead>
            <TableHead>Desc</TableHead>
            <TableHead>Last update</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="rows.length === 0">
            <TableCell colspan="12" class="text-muted-foreground">No agents</TableCell>
          </TableRow>
          <TableRow v-for="a in rows" :key="a.agentId">
            <TableCell>
              <span
                class="inline-block h-2.5 w-2.5 rounded-full"
                :class="light(a.status)"
                :title="a.status"
              />
            </TableCell>
            <TableCell>
              <RouterLink class="underline" :to="`/agents/${encodeURIComponent(a.agentId)}`">
                {{ formatAgentId(a.agentId, idFormat) }}
              </RouterLink>
            </TableCell>
            <TableCell>{{ a.hostname }}</TableCell>
            <TableCell>{{ pct(a.cpuPercent) }}</TableCell>
            <TableCell>{{ pct(a.memoryUsedPercent) }}</TableCell>
            <TableCell>{{ a.disks }}</TableCell>
            <TableCell>{{ a.os }}</TableCell>
            <TableCell>{{ a.ip }}</TableCell>
            <TableCell>{{ a.latencyMs == null ? '—' : `${a.latencyMs} ms` }}</TableCell>
            <TableCell class="max-w-48 truncate" :title="a.description">{{ a.description }}</TableCell>
            <TableCell class="whitespace-nowrap">{{ stamp(a.collectedAt || a.lastSeenAt) }}</TableCell>
            <TableCell>
              <div class="flex gap-1">
                <button type="button" title="Restart" class="rounded-md border border-amber-500/70 p-1 text-amber-400" @click="run(a, 'restart')">
                  <RotateCcw class="h-4 w-4" />
                </button>
                <button type="button" title="Shutdown" class="rounded-md border border-orange-500/70 p-1 text-orange-400" @click="run(a, 'shutdown')">
                  <Power class="h-4 w-4" />
                </button>
                <button type="button" title="Delete" class="rounded-md border border-red-500/70 p-1 text-red-400" @click="run(a, 'delete')">
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
