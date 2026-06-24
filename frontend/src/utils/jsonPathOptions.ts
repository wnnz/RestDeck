import type { JsonPathOption } from '../types'

export function jsonPathOptions(raw: string, limit = 80): JsonPathOption[] {
  if (!raw.trim()) return []
  try {
    const parsed = JSON.parse(raw)
    const out: JsonPathOption[] = []
    walkJSON(parsed, '$', out, limit)
    return out
  } catch {
    return []
  }
}

function walkJSON(value: unknown, path: string, out: JsonPathOption[], limit: number) {
  if (out.length >= limit) return
  if (path !== '$') {
    out.push({ path, label: path, preview: previewValue(value) })
  }
  if (Array.isArray(value)) {
    value.slice(0, 20).forEach((item, index) => walkJSON(item, `${path}[${index}]`, out, limit))
    return
  }
  if (value && typeof value === 'object') {
    for (const key of Object.keys(value as Record<string, unknown>).slice(0, 40)) {
      const childPath = /^[A-Za-z_$][\w$]*$/.test(key)
        ? `${path}.${key}`
        : `${path}[${JSON.stringify(key)}]`
      walkJSON((value as Record<string, unknown>)[key], childPath, out, limit)
      if (out.length >= limit) return
    }
  }
}

function previewValue(value: unknown) {
  if (value == null) return 'null'
  if (typeof value === 'string') return value.length > 80 ? `${value.slice(0, 80)}...` : value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (Array.isArray(value)) return `[${value.length}]`
  return '{...}'
}
