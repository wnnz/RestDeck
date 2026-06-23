import { domain } from '../../wailsjs/go/models'

export function authBadgeCount(request: domain.Request | null) {
  if (!request?.auth?.type || request.auth.type === 'none') return 0
  return 1
}

export function defaultAuthValues(type: string) {
  switch (type) {
    case 'apiKey':
      return { key: 'X-API-Key', value: '', in: 'header' }
    case 'bearer':
      return { token: '' }
    case 'basic':
    case 'digest':
      return { username: '', password: '' }
    case 'oauth1':
      return { consumerKey: '', consumerSecret: '', token: '', tokenSecret: '' }
    case 'oauth2':
      return { accessToken: '' }
    default:
      return {}
  }
}

export function newKeyValue() {
  return new domain.KeyValue({
    id: crypto.randomUUID(),
    enabled: true,
    key: '',
    value: '',
    description: '',
    secret: false,
    valueType: 'static',
    timestampFormat: 'seconds',
    sourceRequestId: '',
    jsonPath: '$.',
    responseStrategy: 'latestHistory',
    refreshAfterSeconds: 300,
    fallbackValue: ''
  })
}

export function newFormItem() {
  return new domain.FormItem({ id: crypto.randomUUID(), enabled: true, key: '', type: 'text', value: '', filePath: '', description: '' })
}

export function cloneKeyValues(items: domain.KeyValue[]) {
  return items.map((item) => normalizeKeyValue(new domain.KeyValue({ ...item })))
}

export function cloneRequest(request: domain.Request) {
  return new domain.Request(JSON.parse(JSON.stringify(request)))
}

export function normalizeRequest(request: domain.Request) {
  request.params = request.params ?? []
  request.headers = request.headers ?? []
  request.params = request.params.map(normalizeKeyValue)
  request.headers = request.headers.map(normalizeKeyValue)
  request.proxy = normalizeProxy(request.proxy, 'inherit')
  request.proxy.noProxy = ''
  if (request.bodyMode === 'form') {
    request.formItems = normalizeFormItems(request.formItems ?? [], request.body ?? '')
    request.body = formItemsToBody(request.formItems)
  } else {
    request.formItems = normalizeFormItems(request.formItems ?? [], '')
  }
  if (!request.auth) {
    request.auth = new domain.AuthConfig({ type: 'none', values: {} })
  }
  return request
}

export function normalizeKeyValue(item: domain.KeyValue) {
  item.id = item.id || crypto.randomUUID()
  item.enabled = item.enabled ?? true
  item.key = item.key ?? ''
  item.value = item.value ?? ''
  item.description = item.description ?? ''
  item.secret = item.secret ?? false
  item.valueType = item.valueType || 'static'
  item.timestampFormat = item.timestampFormat || 'seconds'
  item.sourceRequestId = item.sourceRequestId || ''
  item.jsonPath = item.jsonPath || '$.'
  item.responseStrategy = item.responseStrategy || 'latestHistory'
  item.refreshAfterSeconds = item.refreshAfterSeconds || 300
  item.fallbackValue = item.fallbackValue || ''
  return item
}

export function normalizeProxy(proxy: domain.ProxyConfig | undefined, fallbackMode: 'inherit' | 'none') {
  const mode = proxy?.mode === 'custom' || proxy?.mode === 'none' || proxy?.mode === 'inherit' ? proxy.mode : fallbackMode
  return new domain.ProxyConfig({
    mode,
    url: mode === 'custom' ? (proxy?.url ?? '') : '',
    noProxy: mode === 'custom' ? normalizeNoProxy(proxy?.noProxy ?? '') : ''
  })
}

export function normalizeNoProxy(raw: string) {
  return raw
    .split(/[,\s]+/)
    .map((item) => item.trim())
    .filter(Boolean)
    .filter((item, index, list) => list.indexOf(item) === index)
    .join(',')
}

export function normalizeFormItems(items: domain.FormItem[], fallbackBody: string) {
  const source = items.length ? items : formItemsFromBody(fallbackBody)
  const normalized = source.map((item) => new domain.FormItem({
    id: item.id || crypto.randomUUID(),
    enabled: item.enabled ?? true,
    key: item.key ?? '',
    type: item.type === 'file' ? 'file' : 'text',
    value: item.value ?? '',
    filePath: item.filePath ?? '',
    description: item.description ?? ''
  }))
  return normalized.length ? normalized : [newFormItem()]
}

export function formItemsFromBody(raw: string) {
  return raw
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const index = line.indexOf('=')
      const key = index >= 0 ? line.slice(0, index).trim() : line.trim()
      const value = index >= 0 ? line.slice(index + 1).trim() : ''
      if (value.startsWith('@')) {
        return new domain.FormItem({ id: crypto.randomUUID(), enabled: true, key, type: 'file', value: '', filePath: value.slice(1), description: '' })
      }
      return new domain.FormItem({ id: crypto.randomUUID(), enabled: true, key, type: 'text', value, filePath: '', description: '' })
    })
}

export function formItemsToBody(items: domain.FormItem[]) {
  return (items ?? [])
    .filter((item) => item.key || item.value || item.filePath)
    .map((item) => `${item.key}=${item.type === 'file' ? `@${item.filePath}` : item.value}`)
    .join('\n')
}

export function syncFormBody(request: domain.Request) {
  if (request.bodyMode === 'form') {
    request.formItems = normalizeFormItems(request.formItems ?? [], request.body ?? '')
    request.body = formItemsToBody(request.formItems)
  } else {
    request.formItems = normalizeFormItems(request.formItems ?? [], '')
  }
}
