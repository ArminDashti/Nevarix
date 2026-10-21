<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { core, type ScanHit } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const cidr = ref('')
const port = ref(9000)
const timeoutMs = ref(500)
const scanning = ref(false)
const error = ref('')
const scanned = ref<number | null>(null)
const hits = ref<ScanHit[]>([])
const added = ref<Set<string>>(new Set())
const adding = ref<Set<string>>(new Set())
const addingAll = ref(false)

const pendingHits = computed(() => hits.value.filter((h) => !added.value.has(h.ip)))

onMounted(async () => {
  try {
    const s = await core.settings()
    cidr.value = s.defaultCidr
    port.value = s.agentPort
    timeoutMs.value = s.timeoutMs
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load settings'
  }
})

async function startScan() {
  scanning.value = true
  error.value = ''
  hits.value = []
  scanned.value = null
  added.value = new Set()
  try {
    const res = await core.scan({ cidr: cidr.value, port: port.value, timeoutMs: timeoutMs.value })
    scanned.value = res.scanned
    hits.value = res.hits
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Scan failed'
  } finally {
    scanning.value = false
  }
}

async function addHit(hit: ScanHit) {
  if (added.value.has(hit.ip) || adding.value.has(hit.ip)) return
  adding.value = new Set([...adding.value, hit.ip])
  error.value = ''
  try {
    await core.addEndpoint({ ip: hit.ip, hostname: hit.hostname })
    added.value = new Set([...added.value, hit.ip])
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to add endpoint'
  } finally {
    const next = new Set(adding.value)
    next.delete(hit.ip)
    adding.value = next
  }
}

async function addAllHits() {
  if (!pendingHits.value.length || addingAll.value) return
  addingAll.value = true
  error.value = ''
  for (const hit of pendingHits.value) {
    await addHit(hit)
  }
  addingAll.value = false
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold">Scan network</h2>
        <p class="text-sm text-muted-foreground">
          Probe each host in a CIDR for nevarix-agent on
          <code class="text-xs">/health</code> (max /16).
        </p>
      </div>
      <RouterLink
        to="/endpoints"
        class="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-muted"
      >
        Back to endpoints
      </RouterLink>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Scan for agents</CardTitle>
      </CardHeader>
      <CardContent class="space-y-3">
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="startScan">
          <label class="flex flex-col text-sm">
            CIDR
            <input
              v-model="cidr"
              class="min-w-[10rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="192.168.1.0/24"
              required
            />
          </label>
          <label class="flex flex-col text-sm">
            Agent port
            <input
              v-model.number="port"
              type="number"
              min="1"
              max="65535"
              class="w-28 rounded-md border border-border bg-background px-2 py-1"
            />
          </label>
          <label class="flex flex-col text-sm">
            Timeout (ms)
            <input
              v-model.number="timeoutMs"
              type="number"
              min="50"
              max="10000"
              class="w-28 rounded-md border border-border bg-background px-2 py-1"
            />
          </label>
          <button
            type="submit"
            :disabled="scanning"
            class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
          >
            {{ scanning ? 'Scanning…' : 'Start scan' }}
          </button>
        </form>
        <p class="text-xs text-muted-foreground">
          Defaults come from
          <RouterLink to="/settings" class="text-primary underline">Settings</RouterLink>.
          Large subnets can take several minutes.
        </p>
      </CardContent>
    </Card>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card v-if="scanned !== null">
      <CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2 space-y-0">
        <CardTitle>{{ hits.length }} agent(s) found in {{ scanned }} hosts</CardTitle>
        <button
          v-if="pendingHits.length"
          type="button"
          :disabled="addingAll || scanning"
          class="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground disabled:opacity-50"
          @click="addAllHits"
        >
          {{ addingAll ? 'Adding…' : `Add all (${pendingHits.length})` }}
        </button>
      </CardHeader>
      <CardContent>
        <p v-if="hits.length === 0" class="text-sm text-muted-foreground">
          No agents answered on this port. Check the CIDR, firewall, and that agents listen on the port above.
        </p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Hostname</TableHead>
              <TableHead>Agent ID</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Port</TableHead>
              <TableHead>Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="h in hits" :key="h.ip">
              <TableCell>{{ h.hostname || '—' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ h.agentId || '—' }}</TableCell>
              <TableCell>{{ h.ip }}</TableCell>
              <TableCell>{{ h.port }}</TableCell>
              <TableCell>
                <button
                  v-if="!added.has(h.ip)"
                  type="button"
                  class="text-sm text-primary underline disabled:opacity-50"
                  :disabled="adding.has(h.ip) || addingAll"
                  @click="addHit(h)"
                >
                  {{ adding.has(h.ip) ? 'Adding…' : 'Add' }}
                </button>
                <span v-else class="text-sm text-muted-foreground">Added</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
