import { domain } from '../../wailsjs/go/models'

export function formatError(error: unknown) {
  if (error instanceof Error) return error.message
  return String(error)
}

export function formatBytes(bytes?: number) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

export function statusClass(code?: number) {
  if (!code) return 'text-zinc-500'
  if (code >= 200 && code < 300) return 'text-emerald-600'
  if (code >= 300 && code < 400) return 'text-sky-600'
  if (code >= 400) return 'text-red-600'
  return 'text-zinc-600'
}

export function responseStatusText(item: domain.Response) {
  return item.status || String(item.statusCode || '')
}
