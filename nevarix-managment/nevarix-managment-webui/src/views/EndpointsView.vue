<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { hub, type HubEndpoint } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const endpoints = ref<HubEndpoint[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await hub.endpoints()
    endpoints.value = res.endpoints
    error.value = res.agentError ?? ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 5000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">Endpoints</h2>
      <RouterLink to="/endpoints/scan" class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">
        Scan network
      </RouterLink>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>{{ endpoints.length }} endpoints</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hostname</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Source</TableHead>
              <TableHead>Last seen</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(e, i) in endpoints" :key="e.ip ?? e.hostname + i">
              <TableCell>{{ e.hostname || '-' }}</TableCell>
              <TableCell>{{ e.ip ?? '-' }}</TableCell>
              <TableCell>{{ e.source }}</TableCell>
              <TableCell>{{ e.lastSeenAt ? new Date(e.lastSeenAt).toLocaleString() : '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
