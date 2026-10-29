<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { core, type CoreAction } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const actions = ref<CoreAction[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await core.actions()
    actions.value = res.actions
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
    <h2 class="text-lg font-semibold">Actions</h2>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>{{ actions.length }} actions</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Kind</TableHead>
              <TableHead>Detail</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(a, i) in actions" :key="i">
              <TableCell>{{ new Date(a.createdAt).toLocaleString() }}</TableCell>
              <TableCell>{{ a.kind }}</TableCell>
              <TableCell>{{ a.detail || '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
