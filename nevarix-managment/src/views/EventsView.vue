<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { core, type CoreEvent } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const events = ref<CoreEvent[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await core.events()
    events.value = res.events
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
    <h2 class="text-lg font-semibold">Events</h2>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>{{ events.length }} events</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Agent</TableHead>
              <TableHead>Kind</TableHead>
              <TableHead>Message</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(e, i) in events" :key="i">
              <TableCell>{{ new Date(e.occurredAt).toLocaleString() }}</TableCell>
              <TableCell>{{ e.agentId }}</TableCell>
              <TableCell>{{ e.kind }}</TableCell>
              <TableCell>{{ e.message || '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
