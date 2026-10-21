<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { core, type NetDevice, type NetInterface } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const route = useRoute()
const deviceId = computed(() => route.params.deviceId as string)
const device = ref<NetDevice | null>(null)
const interfaces = ref<NetInterface[]>([])
const error = ref('')

async function load() {
  try {
    const [d, ifaces] = await Promise.all([
      core.device(deviceId.value),
      core.deviceInterfaces(deviceId.value),
    ])
    device.value = d.device
    interfaces.value = ifaces.interfaces
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-lg font-semibold">{{ device?.name || device?.mgmtIp || 'Device' }}</h2>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card v-if="device">
      <CardHeader><CardTitle>Overview</CardTitle></CardHeader>
      <CardContent class="space-y-1 text-sm">
        <div>IP: {{ device.mgmtIp }}</div>
        <div>Vendor: {{ device.vendor }}</div>
        <div v-if="device.sysDescr" class="text-muted-foreground">{{ device.sysDescr }}</div>
        <div v-if="device.lastError" class="text-destructive">{{ device.lastError }}</div>
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>Interfaces</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Index</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Oper</TableHead>
              <TableHead>Rx bps</TableHead>
              <TableHead>Tx bps</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="iface in interfaces" :key="iface.ifIndex">
              <TableCell>{{ iface.ifIndex }}</TableCell>
              <TableCell>{{ iface.name }}</TableCell>
              <TableCell>{{ iface.operStatus === 1 ? 'up' : iface.operStatus }}</TableCell>
              <TableCell>{{ iface.rxBps.toFixed(0) }}</TableCell>
              <TableCell>{{ iface.txBps.toFixed(0) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
