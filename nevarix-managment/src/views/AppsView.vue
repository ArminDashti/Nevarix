<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { core, type CoreApp } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const apps = ref<CoreApp[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await core.apps()
    apps.value = res.apps
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 15000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-lg font-semibold">Apps</h2>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>{{ apps.length }} installed apps</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Version</TableHead>
              <TableHead>Publisher</TableHead>
              <TableHead>Installed</TableHead>
              <TableHead>Source</TableHead>
              <TableHead>Agent</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="a in apps" :key="`${a.agentId}:${a.name}`">
              <TableCell>{{ a.name }}</TableCell>
              <TableCell>{{ a.version || '-' }}</TableCell>
              <TableCell>{{ a.publisher || '-' }}</TableCell>
              <TableCell>{{ a.installDate || '-' }}</TableCell>
              <TableCell>{{ a.source || '-' }}</TableCell>
              <TableCell>{{ a.agentId }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
