<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { core, type SnmpScanHit } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const cidr = ref('')
const timeoutMs = ref(500)
const scanned = ref(0)
const hits = ref<SnmpScanHit[]>([])
const busy = ref(false)
const error = ref('')

onMounted(async () => {
  try {
    const st = await core.settings()
    cidr.value = st.defaultCidr
    timeoutMs.value = st.timeoutMs
  } catch {
    cidr.value = '192.168.1.0/24'
  }
})

async function runScan() {
  busy.value = true
  error.value = ''
  try {
    const res = await core.scanSnmp({ cidr: cidr.value, timeoutMs: timeoutMs.value })
    scanned.value = res.scanned
    hits.value = res.hits
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Scan failed'
  } finally {
    busy.value = false
  }
}

async function addFromHit(hit: SnmpScanHit) {
  await core.addDevice({ mgmtIp: hit.ip, name: hit.vendor + ' ' + hit.ip, snmpCommunity: 'public', pollIntervalSec: 60 })
}
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-lg font-semibold">SNMP device scan</h2>
    <Card>
      <CardContent class="pt-6">
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="runScan">
          <label class="flex flex-col text-sm">
            CIDR
            <input v-model="cidr" required class="rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <label class="flex flex-col text-sm">
            Timeout (ms)
            <input v-model.number="timeoutMs" type="number" min="50" class="w-28 rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <button
            type="submit"
            class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
            :disabled="busy"
          >
            {{ busy ? 'Scanning…' : 'Scan' }}
          </button>
        </form>
        <p v-if="error" class="mt-3 text-sm text-destructive">{{ error }}</p>
      </CardContent>
    </Card>
    <Card v-if="scanned">
      <CardHeader><CardTitle>{{ hits.length }} SNMP hosts (of {{ scanned }} probed)</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>IP</TableHead>
              <TableHead>Vendor</TableHead>
              <TableHead>sysDescr</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="h in hits" :key="h.ip">
              <TableCell>{{ h.ip }}</TableCell>
              <TableCell>{{ h.vendor }}</TableCell>
              <TableCell class="max-w-md truncate text-xs">{{ h.sysDescr }}</TableCell>
              <TableCell>
                <button type="button" class="text-sm text-primary hover:underline" @click="addFromHit(h)">Add</button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
