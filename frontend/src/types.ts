export type NavKey = 'collections' | 'environments' | 'history' | 'runner' | 'realtime' | 'settings'
export type RequestTab = 'params' | 'auth' | 'headers' | 'body' | 'pre' | 'tests' | 'preview' | 'variables' | 'settings'
export type ResponseTab = 'body' | 'headers' | 'cookies' | 'tests'
export type ResponseView = 'pretty' | 'raw' | 'preview'
export type ActiveModal = 'postman' | 'fetch' | 'curl' | 'openapi' | 'swagger' | 'har' | 'export' | null
export type Language = 'zh-CN' | 'en-US'
export type Theme = 'light' | 'dark'
export type ProxyMode = 'inherit' | 'none' | 'custom'
export type VariableValueType = 'static' | 'timestamp' | 'responseJsonPath'
export type TimestampFormat = 'seconds' | 'milliseconds' | 'iso'
export type ResponseStrategy = 'latestHistory' | 'alwaysRequest' | 'refreshAfter'
export type VariableSuggestion = { name: string; detail: string }
export type JsonTokenType = 'plain' | 'key' | 'string' | 'number' | 'boolean' | 'null' | 'punctuation'
export type JsonToken = { type: JsonTokenType; text: string }
export type JsonPathOption = { path: string; label: string; preview: string }
export type RunnerItemStatus = 'waiting' | 'running' | 'passed' | 'failed' | 'skipped'
export type RunnerFailurePolicy = 'continue' | 'stopOnFailure'
export type RunnerQueueItem = {
  id: string
  requestId: string
  iteration: number
  method: string
  name: string
  url: string
  status: RunnerItemStatus
  statusCode?: number
  durationMs?: number
  message?: string
  request?: import('../wailsjs/go/models').domain.PreparedRequest
  response?: import('../wailsjs/go/models').domain.Response
  testResults?: import('../wailsjs/go/models').domain.TestResult[]
}
