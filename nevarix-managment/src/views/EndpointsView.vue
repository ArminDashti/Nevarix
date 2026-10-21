<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { core, type CoreEndpoint } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const endpoints = ref<CoreEndpoint[]>([])
const error = ref('')
const saving = ref(false)
const showAdd = ref(true)
const form = ref({
  ip: '',
  hostname: '',
  agentPort: 9000,
  username: '',
  password: '',
  verifyAgent: false,
})
let timer: number | undefined

async function load() {
  try {
    const res = await core.endpoints()
    endpoints.value = res.endpoints
    error.value = res.agentError ?? ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

async function loadDefaults() {
  try {
    const s = await core.settings()
    form.value.agentPort = s.agentPort
  } catch {
    /* keep default */
  }
}

async function addAgent() {
  saving.value = true
  error.value = ''
  try {
    await core.addEndpoint({
      ip: form.value.ip.trim(),
      hostname: form.value.hostname.trim(),
      agentPort: form.value.agentPort || undefined,
      username: form.value.username.trim() || undefined,
      password: form.value.password || undefined,
      verifyAgent: form.value.verifyAgent,
    })
    form.value = {
      ip: '',
      hostname: '',
      agentPort: form.value.agentPort,
      username: '',
      password: '',
      verifyAgent: false,
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
      <h2 class="text-lg font-semibold">Agent endpoints</h2>
      <RouterLink
        to="/endpoints/scan"
        class="rounded-md border border-border px-3 py-2 text-sm font-medium hover:bg-muted"
      >
        Scan network
      </RouterLink>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card v-if="showAdd">
      <CardHeader><CardTitle>Add agent by IP</CardTitle></CardHeader>
      <CardContent>
        <p class="mb-3 text-sm text-muted-foreground">
          Register a host for inventory. Optional SSH credentials are stored for future access; the agent must already
          be installed unless you enable verify.
        </p>
        <form class="flex flex-wrap items-end gap-3" @submit.prevent="addAgent">
          <label class="flex flex-col text-sm">
            IP address
            <input
              v-model="form.ip"
              required
              class="min-w-[10rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="192.168.1.10"
            />
          </label>
          <label class="flex flex-col text-sm">
            Hostname
            <input
              v-model="form.hostname"
              class="min-w-[10rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="optional"
            />
          </label>
          <label class="flex flex-col text-sm">
            Agent port
            <input
              v-model.number="form.agentPort"
              type="number"
              min="1"
              max="65535"
              class="w-28 rounded-md border border-border bg-background px-2 py-1"
            />
          </label>
          <label class="flex flex-col text-sm">
            SSH username
            <input
              v-model="form.username"
              class="min-w-[8rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="optional"
              autocomplete="username"
            />
          </label>
          <label class="flex flex-col text-sm">
            SSH password
            <input
              v-model="form.password"
              type="password"
              class="min-w-[8rem] rounded-md border border-border bg-background px-2 py-1"
              placeholder="optional"
              autocomplete="new-password"
            />
          </label>
          <label class="flex items-center gap-2 pb-1 text-sm">
            <input v-model="form.verifyAgent" type="checkbox" class="rounded border-border" />
            Verify agent responds
          </label>
          <button
            type="submit"
            :disabled="saving"
            class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
          >
            {{ saving ? 'Saving…' : 'Add agent' }}
          </button>
        </form>
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>{{ endpoints.length }} endpoints</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hostname</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Port</TableHead>
              <TableHead>SSH user</TableHead>
              <TableHead>Source</TableHead>
              <TableHead>Last seen</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="(e, i) in endpoints" :key="e.ip ?? e.hostname + i">
              <TableCell>{{ e.hostname || '-' }}</TableCell>
              <TableCell>{{ e.ip ?? '-' }}</TableCell>
              <TableCell>{{ e.agentPort || '-' }}</TableCell>
              <TableCell>{{ e.username || '-' }}</TableCell>
              <TableCell>{{ e.source }}</TableCell>
              <TableCell>{{ e.lastSeenAt ? new Date(e.lastSeenAt).toLocaleString() : '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
