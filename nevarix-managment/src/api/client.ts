function withBase(path: string): string {
  const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')
  if (!base || base === '/') return path
  return `${base}${path.startsWith('/') ? path : `/${path}`}`
}

async function getJson<T>(path: string): Promise<T> {
  const res = await fetch(withBase(path))
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return (await res.json()) as T
}

export const api = {
  dashboard: () => getJson<Record<string, unknown>>('/api/v1/dashboard'),
  cpu: () => getJson<Record<string, unknown>>('/api/v1/cpu'),
  memory: () => getJson<Record<string, unknown>>('/api/v1/memory'),
  disk: () => getJson<Record<string, unknown>>('/api/v1/disk'),
  bandwidth: () => getJson<Record<string, unknown>>('/api/v1/bandwidth'),
  dockerImages: () => getJson<{ images: unknown[]; error?: string }>('/api/v1/docker/images'),
  dockerContainers: () => getJson<{ stacks: unknown[]; error?: string }>('/api/v1/docker/containers'),
  system: () => getJson<{ hostname: string; system: SystemInfo | null }>('/api/v1/system'),
}

export type SystemInfo = {
  os: string
  platform: string
  platformVersion: string
  kernelVersion: string
  arch: string
  motherboard: { vendor: string; product: string; serial: string }
  firewall: { name: string; enabled: boolean; defaultInbound: string; defaultOutbound: string }[]
  temperatures: { sensor: string; celsius: number }[]
}

// nevarix-core panel routes live on the same /api/v1 prefix as metrics (Vite/nginx proxy in dev).
function coreUrl(path: string): string {
  const direct = (import.meta.env.VITE_CORE_URL || '').replace(/\/$/, '')
  if (direct) return `${direct}${path}`
  return withBase(path)
}

async function coreJson<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(coreUrl(path), {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  const body = await res.json()
  if (!res.ok) {
    throw new Error(body.error ?? `${res.status} ${res.statusText}`)
  }
  return body as T
}

export const core = {
  url: coreUrl('/api/v1'),
  endpoints: () => coreJson<{ endpoints: CoreEndpoint[]; agentError?: string }>('/api/v1/endpoints'),
  addEndpoint: (body: {
    ip: string
    hostname?: string
    agentPort?: number
    username?: string
    password?: string
    verifyAgent?: boolean
  }) =>
    coreJson<{ status: string }>('/api/v1/endpoints', { method: 'POST', body: JSON.stringify(body) }),
  scan: (body: { cidr: string; port: number; timeoutMs: number }) =>
    coreJson<{ scanned: number; hits: ScanHit[] }>('/api/v1/scan', { method: 'POST', body: JSON.stringify(body) }),
  settings: () => coreJson<CoreSettings>('/api/v1/settings'),
  saveSettings: (settings: CoreSettings) =>
    coreJson<CoreSettings>('/api/v1/settings', { method: 'PUT', body: JSON.stringify(settings) }),
  devices: () => coreJson<{ devices: NetDeviceSummary[] }>('/api/v1/devices'),
  device: (id: string) => coreJson<{ device: NetDevice }>(`/api/v1/devices/${id}`),
  addDevice: (body: {
    name?: string
    mgmtIp: string
    snmpCommunity?: string
    pollIntervalSec?: number
    enabled?: boolean
    sshUsername?: string
    sshPassword?: string
    verifySnmp?: boolean
  }) => coreJson<{ id: string }>('/api/v1/devices', { method: 'POST', body: JSON.stringify(body) }),
  scanSnmp: (body: { cidr: string; timeoutMs: number }) =>
    coreJson<{ scanned: number; hits: SnmpScanHit[] }>('/api/v1/devices/scan', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  deviceInterfaces: (id: string) =>
    coreJson<{ interfaces: NetInterface[] }>(`/api/v1/devices/${id}/interfaces`),
}

export type CoreEndpoint = {
  ip?: string
  hostname: string
  agentPort?: number
  username?: string
  source: 'scan' | 'agent'
  lastSeenAt?: string
}

export type ScanHit = { ip: string; port: number; agentId: string; hostname: string }

export type CoreSettings = {
  defaultCidr: string
  agentPort: number
  timeoutMs: number
  snmpCommunity: string
  netPollIntervalSec: number
}

export type NetDeviceSummary = {
  id: string
  name: string
  mgmtIp: string
  vendor: string
  enabled: boolean
  cpuPercent: number
  memoryUsedPercent: number
  lastOkAt?: string
  lastError?: string
}

export type NetDevice = NetDeviceSummary & {
  snmpVersion: string
  snmpCommunity?: string
  pollIntervalSec: number
  sysDescr?: string
  sysObjectId?: string
  createdAt: string
}

export type NetInterface = {
  ifIndex: number
  name: string
  operStatus: number
  speedBps: number
  rxBps: number
  txBps: number
  sampledAt?: string
}

export type SnmpScanHit = { ip: string; sysDescr: string; vendor: string }
