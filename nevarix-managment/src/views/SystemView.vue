<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, type SystemInfo } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const hostname = ref('')
const system = ref<SystemInfo | null>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await api.system()
    hostname.value = res.hostname
    system.value = res.system
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
      <CardHeader><CardTitle>System info {{ hostname ? `- ${hostname}` : '' }}</CardTitle></CardHeader>
      <CardContent>
        <dl class="grid grid-cols-2 gap-2 text-sm">
          <dt class="text-muted-foreground">OS</dt><dd>{{ system?.os ?? '-' }}</dd>
          <dt class="text-muted-foreground">Platform</dt><dd>{{ system?.platform ?? '-' }}</dd>
          <dt class="text-muted-foreground">Version</dt><dd>{{ system?.platformVersion ?? '-' }}</dd>
          <dt class="text-muted-foreground">Kernel</dt><dd>{{ system?.kernelVersion ?? '-' }}</dd>
          <dt class="text-muted-foreground">Architecture</dt><dd>{{ system?.arch ?? '-' }}</dd>
        </dl>
      </CardContent>
    </Card>
  </div>
</template>
