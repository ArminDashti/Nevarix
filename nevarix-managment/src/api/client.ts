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

// nevarix-core owns endpoints, scan and settings. The webui talks only to core.
const CORE_URL = (import.meta.env.VITE_CORE_URL || 'http://127.0.0.1:8090').replace(/\/$/, '')

async function coreJson<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${CORE_URL}${path}`, {
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
  url: CORE_URL,
  endpoints: () => coreJson<{ endpoints: CoreEndpoint[]; agentError?: string }>('/api/v1/endpoints'),
  addEndpoint: (ip: string, hostname: string) =>
    coreJson<{ status: string }>('/api/v1/endpoints', { method: 'POST', body: JSON.stringify({ ip, hostname }) }),
  scan: (body: { cidr: string; port: number; timeoutMs: number }) =>
    coreJson<{ scanned: number; hits: ScanHit[] }>('/api/v1/scan', { method: 'POST', body: JSON.stringify(body) }),
  settings: () => coreJson<CoreSettings>('/api/v1/settings'),
  saveSettings: (settings: CoreSettings) =>
    coreJson<CoreSettings>('/api/v1/settings', { method: 'PUT', body: JSON.stringify(settings) }),
}

export type CoreEndpoint = {
  ip?: string
  hostname: string
  source: 'scan' | 'agent'
  lastSeenAt?: string
}

export type ScanHit = { ip: string; port: number; agentId: string; hostname: string }

export type CoreSettings = {
  defaultCidr: string
  agentPort: number
  timeoutMs: number
}
