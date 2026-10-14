<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { hub, type HubSettings } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const settings = ref<HubSettings>({ defaultCidr: '', agentPort: 9000, timeoutMs: 500 })
const status = ref('')
const error = ref('')

onMounted(async () => {
  try {
    settings.value = await hub.settings()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load settings'
  }
})

async function save() {
  error.value = ''
  status.value = ''
  try {
    settings.value = await hub.saveSettings(settings.value)
    status.value = 'Saved'
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to save'
  }
}
</script>

<template>
  <div class="space-y-4">
    <Card>
      <CardHeader><CardTitle>Scan defaults</CardTitle></CardHeader>
      <CardContent>
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="save">
          <label class="flex flex-col text-sm">
            Default CIDR
            <input v-model="settings.defaultCidr" class="rounded-md border border-border bg-background px-2 py-1" required />
          </label>
          <label class="flex flex-col text-sm">
            Agent port
            <input v-model.number="settings.agentPort" type="number" min="1" max="65535" class="w-28 rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <label class="flex flex-col text-sm">
            Timeout (ms)
            <input v-model.number="settings.timeoutMs" type="number" min="50" max="10000" class="w-28 rounded-md border border-border bg-background px-2 py-1" />
          </label>
          <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">Save</button>
        </form>
        <p v-if="status" class="mt-3 text-sm text-muted-foreground">{{ status }}</p>
        <p v-if="error" class="mt-3 text-sm text-destructive">{{ error }}</p>
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>Connection</CardTitle></CardHeader>
      <CardContent class="text-sm">Hub API: {{ hub.url }}</CardContent>
    </Card>
  </div>
</template>
