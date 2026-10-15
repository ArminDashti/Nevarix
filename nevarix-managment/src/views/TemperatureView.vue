<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, type SystemInfo } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const system = ref<SystemInfo | null>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    system.value = (await api.system()).system
    error.value = ''
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
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>Temperature sensors</CardTitle></CardHeader>
      <CardContent>
        <p v-if="system && system.temperatures.length === 0" class="text-sm text-muted-foreground">
          No sensors reported by this host.
        </p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Sensor</TableHead>
              <TableHead>Temperature</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="t in system?.temperatures ?? []" :key="t.sensor">
              <TableCell>{{ t.sensor }}</TableCell>
              <TableCell>{{ t.celsius.toFixed(1) }} °C</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
