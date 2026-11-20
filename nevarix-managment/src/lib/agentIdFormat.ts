export type AgentIdFormat = 'full' | 'short' | 'upper'

const key = 'nevarix.agentIdFormat'

export function loadAgentIdFormat(): AgentIdFormat {
  const v = localStorage.getItem(key)
  if (v === 'short' || v === 'upper' || v === 'full') return v
  return 'full'
}

export function saveAgentIdFormat(format: AgentIdFormat) {
  localStorage.setItem(key, format)
}

export function formatAgentId(id: string, format: AgentIdFormat): string {
  if (format === 'short') return id.slice(0, 8)
  if (format === 'upper') return id.toUpperCase()
  return id
}
