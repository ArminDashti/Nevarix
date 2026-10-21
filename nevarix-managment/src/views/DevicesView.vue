<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { core, type NetDeviceSummary } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const devices = ref<NetDeviceSummary[]>([])
const error = ref('')
const saving = ref(false)
const form = ref({
  name: '',
  mgmtIp: '',
  snmpCommunity: 'public',
  pollIntervalSec: 60,
  sshUsername: '',
  sshPassword: '',
  verifySnmp: true,
})
let timer: number | undefined

async function load() {
  try {
    const res = await core.devices()
    devices.value = res.devices
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

async function loadDefaults() {
  try {
    const s = await core.settings()
    form.value.snmpCommunity = s.snmpCommunity
    form.value.pollIntervalSec = s.netPollIntervalSec
  } catch {
    /* keep defaults */
  }
}

async function addDevice() {
  saving.value = true
  error.value = ''
  try {
    await core.addDevice({
      name: form.value.name.trim() || undefined,
      mgmtIp: form.value.mgmtIp.trim(),
      snmpCommunity: form.value.snmpCommunity,
      pollIntervalSec: form.value.pollIntervalSec,
      sshUsername: form.value.sshUsername.trim() || undefined,
      sshPassword: form.value.sshPassword || undefined,
      verifySnmp: form.value.verifySnmp,
    })
    form.value = {
      name: '',
      mgmtIp: '',
      snmpCommunity: form.value.snmpCommunity,
      pollIntervalSec: form.value.pollIntervalSec,
      sshUsername: '',
      sshPassword: '',
      verifySnmp: true,
    }
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to add'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadDefaults()
  load()
  timer = window.setInterval(load, 5000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-lg font-semibold">SNMP network devices</h2>
      <RouterLink
        to="/devices/scan"
        class="rounded-md border border-border px-3 py-2 text-sm font-medium hover:bg-muted"
      >
        SNMP scan
      </RouterLink>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>Add device by IP (SNMP v2c)</CardTitle></CardHeader>
      <CardContent>
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="addDevice">
          <label class="flex flex-col text-sm">
            Management IP
            <input
              v-model="form.mgmtIp"
              required
              class="min-w-[10rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="10.0.0.1"
            />
          </label>
          <label class="flex flex-col text-sm">
            Name
            <input v-model="form.name" class="rounded-md border border-border bg-background px-2 py-1" placeholder="optional" />
          </label>
          <label class="flex flex-col text-sm">
            SNMP community
            <input
              v-model="form.snmpCommunity"
              type="password"
              class="min-w-[8rem] rounded-md border border-border bg-background px-2 py-1"
              autocomplete="off"
            />
          </label>
          <label class="flex flex-col text-sm">
            Poll interval (s)
            <input
              v-model.number="form.pollIntervalSec"
              type="number"
              min="10"
              class="w-24 rounded-md border border-border bg-background px-2 py-1"
            />
          </label>
          <label class="flex flex-col text-sm">
            SSH username
            <input
              v-model="form.sshUsername"
              class="min-w-[8rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="optional"
              autocomplete="username"
            />
          </label>
          <label class="flex flex-col text-sm">
            SSH password
            <input
              v-model="form.sshPassword"
              type="password"
              class="min-w-[8rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="optional"
              autocomplete="new-password"
            />
          </label>
          <label class="flex items-center gap-2 pb-1 text-sm">
            <input v-model="form.verifySnmp" type="checkbox" class="rounded border-border" />
            Verify SNMP before save
          </label>
          <button
            type="submit"
            :disabled="saving"
            class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
          >
            {{ saving ? 'Saving…' : 'Add device' }}
          </button>
        </form>
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>{{ devices.length }} devices</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Vendor</TableHead>
              <TableHead>CPU %</TableHead>
              <TableHead>Mem %</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="d in devices" :key="d.id">
              <TableCell>
                <RouterLink :to="`/devices/${d.id}`" class="font-medium text-primary hover:underline">
                  {{ d.name || d.mgmtIp }}
                </RouterLink>
              </TableCell>
              <TableCell>{{ d.mgmtIp }}</TableCell>
              <TableCell>{{ d.vendor }}</TableCell>
              <TableCell>{{ d.cpuPercent.toFixed(1) }}</TableCell>
              <TableCell>{{ d.memoryUsedPercent.toFixed(1) }}</TableCell>
              <TableCell class="text-xs">
                <span v-if="d.lastError" class="text-destructive">{{ d.lastError }}</span>
                <span v-else-if="d.lastOkAt">OK {{ new Date(d.lastOkAt).toLocaleString() }}</span>
                <span v-else class="text-muted-foreground">waiting</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
