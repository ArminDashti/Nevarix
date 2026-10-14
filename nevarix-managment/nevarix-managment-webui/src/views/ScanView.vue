<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { hub, type ScanHit } from '@/api/client'
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

onMounted(async () => {
  try {
    const s = await hub.settings()
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
  try {
    const res = await hub.scan({ cidr: cidr.value, port: port.value, timeoutMs: timeoutMs.value })
    scanned.value = res.scanned
    hits.value = res.hits
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Scan failed'
  } finally {
    scanning.value = false
  }
}

async function addHit(hit: ScanHit) {
  try {
    await hub.addEndpoint(hit.ip, hit.hostname)
    added.value = new Set([...added.value, hit.ip])
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to add endpoint'
  }
}
</script>

<template>
  <div class="space-y-4">
    <Card>
      <CardHeader><CardTitle>Scan for agents</CardTitle></CardHeader>
      <CardContent>
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="startScan">
          <label class="flex flex-col text-sm">
            CIDR
            <input v-model="cidr" class="rounded-md border border-border bg-background px-2 py-1" required />
          </label>
          <label class="flex flex-col text-sm">
            Agent port
            <input v-model.number="port" type="number" min="1" max="65535" class="w-28 rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <label class="flex flex-col text-sm">
            Timeout (ms)
            <input v-model.number="timeoutMs" type="number" min="50" max="10000" class="w-28 rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <button type="submit" :disabled="scanning" class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50">
            {{ scanning ? 'Scanning...' : 'Start scan' }}
          </button>
        </form>
      </CardContent>
    </Card>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card v-if="scanned !== null">
      <CardHeader><CardTitle>{{ hits.length }} agent(s) found in {{ scanned }} hosts</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>IP</TableHead>
              <TableHead>Port</TableHead>
              <TableHead>Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="h in hits" :key="h.ip">
              <TableCell>{{ h.ip }}</TableCell>
              <TableCell>{{ h.port }}</TableCell>
              <TableCell>
                <button v-if="!added.has(h.ip)" class="text-sm text-primary underline" @click="addHit(h)">Add</button>
                <span v-else class="text-sm text-muted-foreground">Added</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
