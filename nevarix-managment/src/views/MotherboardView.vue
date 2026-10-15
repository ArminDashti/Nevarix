<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, type SystemInfo } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

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
      <CardHeader><CardTitle>Motherboard</CardTitle></CardHeader>
      <CardContent>
        <dl class="grid grid-cols-2 gap-2 text-sm">
          <dt class="text-muted-foreground">Vendor</dt><dd>{{ system?.motherboard.vendor || '-' }}</dd>
          <dt class="text-muted-foreground">Product</dt><dd>{{ system?.motherboard.product || '-' }}</dd>
          <dt class="text-muted-foreground">Serial</dt><dd>{{ system?.motherboard.serial || '-' }}</dd>
        </dl>
      </CardContent>
    </Card>
  </div>
</template>
