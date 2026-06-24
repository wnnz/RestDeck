<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import {
  Globe2,
  History,
  ListTree,
  Play,
  Settings
} from 'lucide-vue-next'
import { ClipboardSetText, Quit, WindowMinimise, WindowToggleMaximise } from '../wailsjs/runtime/runtime'
import {
  ClearCookies,
  CreateCollection,
  CreateResponseVariable,
  CreateEnvironment,
  DebugVariables,
  DebugRequestVariables,
  DeleteCollection,
  DeleteCookie,
  DeleteEnvironment,
  DeleteRequest,
  ExportHARCollection,
  ExportOpenAPICollection,
  ExportPostmanCollection,
  ExportPostmanRequest,
  FormatBody,
  GetState,
  ImportCurlRequest,
  ImportFetchRequest,
  ImportHARCollection,
  ImportOpenAPICollection,
  ImportOpenAPICollectionWithOptions,
  ImportPostmanCollection,
  ImportSwaggerURL,
  InspectOpenAPI,
  OpenTextFile,
  PreviewRequest,
  QueryJSONPath,
  SaveRunnerResult,
  SaveCollection,
  SaveEnvironment,
  SaveGlobals,
  SaveRequest,
  SaveSettings,
  SaveTextFile,
  SelectFile,
  SendRequest,
  SetActiveEnvironment,
  TestVariable,
  TestSSE,
  TestWebSocket
} from '../wailsjs/go/main/App'
import { domain, realtime } from '../wailsjs/go/models'
import AppTitleBar from './components/AppTitleBar.vue'
import EnvironmentsView from './components/EnvironmentsView.vue'
import HistoryView from './components/HistoryView.vue'
import ImportModal from './components/ImportModal.vue'
import RealtimeView from './components/RealtimeView.vue'
import RequestCodeModal from './components/RequestCodeModal.vue'
import RequestWorkspace from './components/RequestWorkspace.vue'
import RunnerView from './components/RunnerView.vue'
import SettingsView from './components/SettingsView.vue'
import SidebarRail from './components/SidebarRail.vue'
import WorkspaceSidebar from './components/WorkspaceSidebar.vue'
import { useRunnerController } from './composables/useRunnerController'
import { authTypes, bodyModes, methods } from './constants/request'
import { messages } from './i18n/messages'
import type { ActiveModal, JsonToken, Language, NavKey, RequestTab, ResponseTab, ResponseView, Theme, VariableSuggestion } from './types'
import { formatError } from './utils/format'
import { tokenizeJSON } from './utils/jsonHighlight'
import { jsonPathOptions } from './utils/jsonPathOptions'
import {
  authBadgeCount,
  cloneKeyValues,
  cloneRequest,
  defaultAuthValues,
  newFormItem,
  newKeyValue,
  normalizeKeyValue,
  normalizeProxy,
  normalizeRequest,
  syncFormBody
} from './utils/requestModel'

const state = ref<domain.WorkspaceState | null>(null)
const language = ref<Language>((localStorage.getItem('restdeck.language') as Language) || 'zh-CN')
const theme = ref<Theme>((localStorage.getItem('restdeck.theme') as Theme) || 'light')
const activeNav = ref<NavKey>('collections')
const environmentPanel = ref<'environment' | 'globals'>('environment')
const activeRequestTab = ref<RequestTab>('params')
const activeResponseTab = ref<ResponseTab>('body')
const responseView = ref<ResponseView>('pretty')
const activeCollectionId = ref('')
const activeRequest = ref<domain.Request | null>(null)
const response = ref<domain.Response | null>(null)
const search = ref('')
const busy = ref(false)
const realtimeBusy = ref(false)
const statusMessage = ref('')
const activeModal = ref<ActiveModal>(null)
const postmanText = ref('')
const openAPIText = ref('')
const swaggerUrl = ref('')
const harText = ref('')
const fetchText = ref('')
const curlText = ref('')
const collectionPickerOpen = ref(false)
const optionsMenuOpen = ref(false)
const editingCollectionId = ref('')
const editingCollectionName = ref('')
const pendingDeleteCollectionId = ref('')
const exportText = ref('')
const codeModalRequest = ref<domain.Request | null>(null)
const responseSearch = ref('')
const responseJSONPath = ref('$.')
const responseJSONPathResult = ref('')
const responseVariableKey = ref('')
const requestPreview = ref<domain.PreparedRequest | null>(null)
const requestPreviewBusy = ref(false)
const variableDebugReport = ref<domain.VariableDebugReport | null>(null)
const variableDebugBusy = ref(false)
const openAPIServers = ref<string[]>([])
const selectedOpenAPIServer = ref('')
const exportFilename = ref('restdeck-export.json')
const wsDraft = reactive({
  url: 'wss://echo.websocket.events',
  message: '{ "hello": "restdeck" }',
  headers: [] as domain.KeyValue[],
  proxy: new domain.ProxyConfig({ mode: 'inherit', url: '', noProxy: '' }),
  timeoutMs: 10000
})
const sseDraft = reactive({
  url: '{{baseUrl}}/sse',
  headers: [] as domain.KeyValue[],
  proxy: new domain.ProxyConfig({ mode: 'inherit', url: '', noProxy: '' }),
  timeoutMs: 10000,
  maxEvents: 5
})
const wsResult = ref<realtime.WebSocketResult | null>(null)
const sseResult = ref<realtime.SSEResult | null>(null)
let requestAutosaveTimer: ReturnType<typeof setTimeout> | null = null
let environmentAutosaveTimer: ReturnType<typeof setTimeout> | null = null
let globalsAutosaveTimer: ReturnType<typeof setTimeout> | null = null
let settingsAutosaveTimer: ReturnType<typeof setTimeout> | null = null
let lastRequestAutosaveSnapshot = ''
let lastRequestAutosaveId = ''
let lastEnvironmentAutosaveSnapshot = ''
let lastGlobalsAutosaveSnapshot = ''
let lastSettingsAutosaveSnapshot = ''
let suppressRequestAutosave = false
let suppressEnvironmentAutosave = false
let suppressGlobalsAutosave = false
let suppressSettingsAutosave = false
let requestAutosaveQueued = false
let environmentAutosaveQueued = false
let globalsAutosaveQueued = false
let settingsAutosaveQueued = false
let requestAutosavePromise: Promise<void> | null = null
let environmentAutosavePromise: Promise<void> | null = null
let globalsAutosavePromise: Promise<void> | null = null
let settingsAutosavePromise: Promise<void> | null = null
let openAPIInspectTimer: ReturnType<typeof setTimeout> | null = null

const envDraft = reactive({
  id: '',
  name: '',
  variables: [] as domain.KeyValue[]
})
const globalsDraft = ref<domain.KeyValue[]>([])
const settingsDraft = reactive<domain.Settings>(new domain.Settings({ language: 'zh-CN', theme: 'light', defaultProxy: { mode: 'none', url: '', noProxy: '' } }))
const t = computed(() => messages[language.value])
const navItems = computed(() => [
  { key: 'collections' as NavKey, label: t.value.collections, icon: ListTree },
  { key: 'environments' as NavKey, label: t.value.environments, icon: Globe2 },
  { key: 'history' as NavKey, label: t.value.history, icon: History },
  { key: 'runner' as NavKey, label: t.value.runner, icon: Play },
  { key: 'settings' as NavKey, label: t.value.settings, icon: Settings }
])

const activeCollection = computed(() => {
  return state.value?.collections?.find((item) => item.id === activeCollectionId.value) ?? state.value?.collections?.[0] ?? null
})

const activeModalTitle = computed(() => {
  switch (activeModal.value) {
    case 'fetch':
      return t.value.importFetchRequest
    case 'curl':
      return t.value.importCurlRequest
    case 'openapi':
      return t.value.importOpenAPICollection
    case 'swagger':
      return t.value.importSwaggerCollection
    case 'har':
      return t.value.importHARCollection
    case 'postman':
      return t.value.importPostmanCollection
    case 'export':
      return t.value.exportedCollection
    default:
      return ''
  }
})

const activeEnvironment = computed(() => {
  const envs = state.value?.environments ?? []
  return envs.find((item) => item.id === state.value?.activeEnvironmentId) ?? envs.find((item) => item.isActive) ?? envs[0] ?? null
})

const {
  runnerBusy,
  runnerResult,
  runnerScope,
  runnerIterations,
  runnerRequestId,
  runnerQueue,
  runnerFailurePolicy,
  runnerStopRequested,
  runnerRetryCount,
  runnerDelayMs,
  ensureRunnerRequest,
  setRunnerRequest,
  selectRunnerRequest,
  setRunnerScope,
  setRunnerIterations,
  setRunnerFailurePolicy,
  setRunnerRetryCount,
  setRunnerDelayMs,
  stopRunner,
  runActiveCollection,
  runRunnerRequest,
  runnerReportText
} = useRunnerController({
  activeCollection,
  activeEnvironment,
  globalsDraft,
  labels: t,
  statusMessage,
  saveRequest: SaveRequest,
  sendRequest: SendRequest,
  getState: GetState,
  setState,
  saveRunnerResult: SaveRunnerResult
})

const filteredRequests = computed(() => {
  const requests = activeCollection.value?.requests ?? []
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return requests
  return requests.filter((request) => {
    return request.name.toLowerCase().includes(keyword) || request.url.toLowerCase().includes(keyword) || request.method.toLowerCase().includes(keyword)
  })
})

const prettyResponseBody = computed(() => {
  if (!response.value) return ''
  if (responseView.value === 'raw') return response.value.body
  if (responseView.value === 'preview') return response.value.body
  try {
    return JSON.stringify(JSON.parse(response.value.body), null, 2)
  } catch {
    return response.value.body
  }
})

const highlightedResponseBody = computed<JsonToken[]>(() => {
  if (responseView.value !== 'pretty') {
    return [{ type: 'plain', text: prettyResponseBody.value }]
  }
  return tokenizeJSON(prettyResponseBody.value)
})

const responseSearchMatches = computed(() => {
  const keyword = responseSearch.value.trim().toLowerCase()
  if (!keyword || !prettyResponseBody.value) return 0
  return prettyResponseBody.value.toLowerCase().split(keyword).length - 1
})

const responseJSONPathOptions = computed(() => jsonPathOptions(response.value?.body ?? ''))

const requestTabs = computed(() => [
  { key: 'params' as RequestTab, label: t.value.params, count: activeRequest.value?.params?.filter((item) => item.enabled && item.key).length ?? 0 },
  { key: 'auth' as RequestTab, label: t.value.auth, count: authBadgeCount(activeRequest.value) },
  { key: 'headers' as RequestTab, label: t.value.headers, count: activeRequest.value?.headers?.filter((item) => item.enabled && item.key).length ?? 0 },
  { key: 'body' as RequestTab, label: t.value.body, count: activeRequest.value?.bodyMode && activeRequest.value.bodyMode !== 'none' ? 1 : 0 },
  { key: 'pre' as RequestTab, label: t.value.pre, count: activeRequest.value?.preScript?.trim() ? 1 : 0 },
  { key: 'tests' as RequestTab, label: t.value.tests, count: activeRequest.value?.testScript?.trim() ? 1 : 0 },
  { key: 'preview' as RequestTab, label: t.value.actualRequest, count: requestPreview.value?.url ? 1 : 0 },
  { key: 'variables' as RequestTab, label: t.value.variableDebug, count: variableDebugReport.value?.errors?.length ?? 0 },
  { key: 'settings' as RequestTab, label: t.value.settings, count: activeRequest.value?.timeoutMs ? 1 : 0 }
])

const responseTabs = computed(() => [
  { key: 'body' as ResponseTab, label: t.value.body, count: response.value?.body ? 1 : 0 },
  { key: 'headers' as ResponseTab, label: t.value.headers, count: response.value?.headers?.length ?? 0 },
  { key: 'cookies' as ResponseTab, label: t.value.cookies, count: response.value?.cookies?.length ?? 0 },
  { key: 'tests' as ResponseTab, label: t.value.tests, count: response.value?.testResults?.length ?? 0 }
])

const dynamicVariables = [
  { name: '$guid', detail: 'GUID' },
  { name: '$randomUUID', detail: 'UUID' },
  { name: '$timestamp', detail: 'Unix timestamp' },
  { name: '$isoTimestamp', detail: 'ISO timestamp' },
  { name: '$randomInt', detail: 'Random number' },
  { name: '$randomBoolean', detail: 'Random boolean' },
  { name: '$randomEmail', detail: 'Random email' },
  { name: '$randomUserName', detail: 'Random user' },
  { name: '$randomFirstName', detail: 'Random name' }
]

const variableSuggestions = computed<VariableSuggestion[]>(() => {
  const seen = new Set<string>()
  const out: VariableSuggestion[] = []
  const push = (name: string, detail: string) => {
    if (!name || seen.has(name)) return
    seen.add(name)
    out.push({ name, detail })
  }
  for (const variable of globalsDraft.value ?? []) {
    if (variable.enabled && variable.key) push(variable.key, t.value.globalVariable)
  }
  for (const variable of envDraft.variables ?? []) {
    if (variable.enabled && variable.key) push(variable.key, variable.valueType === 'responseJsonPath' ? t.value.responseVariable : t.value.environmentVariable)
  }
  for (const variable of dynamicVariables) {
    push(variable.name, variable.detail)
  }
  return out
})

onMounted(async () => {
  settingsDraft.language = language.value
  settingsDraft.theme = theme.value
  await loadState()
})

watch(language, (next) => {
  settingsDraft.language = next
  localStorage.setItem('restdeck.language', next)
  document.documentElement.lang = next
}, { immediate: true })

watch(theme, (next) => {
  settingsDraft.theme = next
  localStorage.setItem('restdeck.theme', next)
  document.documentElement.dataset.theme = next
}, { immediate: true })

watch(activeEnvironment, (env) => {
  if (!env) return
  suppressEnvironmentAutosave = true
  if (envDraft.id === env.id) {
    envDraft.name = env.name
  } else {
    envDraft.id = env.id
    envDraft.name = env.name
    envDraft.variables = cloneKeyValues(env.variables ?? [])
  }
  lastEnvironmentAutosaveSnapshot = environmentDraftSnapshot()
  void nextTick(() => {
    suppressEnvironmentAutosave = false
  })
}, { immediate: true })

watch(state, (next) => {
  const nextGlobals = cloneKeyValues(next?.globals ?? [])
  const nextGlobalsSnapshot = JSON.stringify(normalizedKeyValues(nextGlobals))
  if (globalsSnapshot() !== nextGlobalsSnapshot) {
    suppressGlobalsAutosave = true
    globalsDraft.value = nextGlobals
    void nextTick(() => {
      suppressGlobalsAutosave = false
    })
  }
  lastGlobalsAutosaveSnapshot = nextGlobalsSnapshot
  if (next?.settings) {
    suppressSettingsAutosave = true
    settingsDraft.language = (next.settings.language as Language) || language.value
    settingsDraft.theme = (next.settings.theme as Theme) || theme.value
    settingsDraft.defaultProxy = normalizeProxy(next.settings.defaultProxy, 'none')
    language.value = settingsDraft.language as Language
    theme.value = settingsDraft.theme as Theme
    lastSettingsAutosaveSnapshot = settingsSnapshot()
    void nextTick(() => {
      suppressSettingsAutosave = false
    })
  }
})

watch(() => activeRequest.value?.id ?? '', (id) => {
  clearRequestAutosaveTimer()
  lastRequestAutosaveId = id
  lastRequestAutosaveSnapshot = requestSnapshot(activeRequest.value)
})

watch(activeRequest, () => {
  scheduleRequestAutosave()
}, { deep: true })

watch(activeRequestTab, (tab) => {
  if (tab === 'preview') void refreshRequestPreview()
  if (tab === 'variables') void refreshVariableDebug()
})

watch(responseJSONPath, () => {
  responseJSONPathResult.value = ''
})

watch(envDraft, () => {
  if (environmentPanel.value === 'environment') scheduleEnvironmentAutosave()
}, { deep: true })

watch(globalsDraft, () => {
  scheduleGlobalsAutosave()
}, { deep: true })

watch(settingsDraft, () => {
  scheduleSettingsAutosave()
}, { deep: true })

watch(openAPIText, () => {
  if (activeModal.value !== 'openapi' && activeModal.value !== 'swagger') return
  if (openAPIInspectTimer) clearTimeout(openAPIInspectTimer)
  openAPIInspectTimer = setTimeout(() => { void inspectOpenAPIText() }, 350)
})

watch(activeNav, (_next, previous) => {
  if (previous === 'collections') void flushRequestAutosave()
  if (previous === 'environments') void flushCurrentEnvironmentPanelAutosave()
  if (previous === 'settings') void flushSettingsAutosave()
})

function clearRequestAutosaveTimer() {
  if (requestAutosaveTimer) {
    clearTimeout(requestAutosaveTimer)
    requestAutosaveTimer = null
  }
}

function clearEnvironmentAutosaveTimer() {
  if (environmentAutosaveTimer) {
    clearTimeout(environmentAutosaveTimer)
    environmentAutosaveTimer = null
  }
}

function clearGlobalsAutosaveTimer() {
  if (globalsAutosaveTimer) {
    clearTimeout(globalsAutosaveTimer)
    globalsAutosaveTimer = null
  }
}

function clearSettingsAutosaveTimer() {
  if (settingsAutosaveTimer) {
    clearTimeout(settingsAutosaveTimer)
    settingsAutosaveTimer = null
  }
}

function requestSnapshot(request: domain.Request | null) {
  if (!request?.id) return ''
  const next = normalizeRequest(cloneRequest(request))
  syncFormBody(next)
  next.updatedAt = ''
  return JSON.stringify(next)
}

function normalizedKeyValues(items: domain.KeyValue[]) {
  return (items ?? []).map((item) => normalizeKeyValue(new domain.KeyValue({ ...item })))
}

function environmentDraftModel() {
  return new domain.Environment({
    id: envDraft.id,
    name: envDraft.name || t.value.environments,
    variables: normalizedKeyValues(envDraft.variables),
    isActive: true
  })
}

function environmentDraftSnapshot() {
  if (!envDraft.id) return ''
  const env = environmentDraftModel()
  env.updatedAt = ''
  return JSON.stringify(env)
}

function globalsSnapshot() {
  return JSON.stringify(normalizedKeyValues(globalsDraft.value))
}

function settingsModel() {
  return new domain.Settings({
    ...settingsDraft,
    language: language.value,
    theme: theme.value,
    defaultProxy: normalizeProxy(settingsDraft.defaultProxy, 'none')
  })
}

function settingsSnapshot() {
  return JSON.stringify(settingsModel())
}

function updateRequestInState(request: domain.Request) {
  if (!state.value?.collections?.length) return
  state.value = new domain.WorkspaceState({
    ...state.value,
    collections: state.value.collections.map((collection) => {
      if (collection.id !== request.collectionId) return collection
      const requests = (collection.requests ?? []).map((item) => item.id === request.id ? normalizeRequest(cloneRequest(request)) : item)
      return new domain.Collection({ ...collection, requests })
    })
  })
}

function updateEnvironmentInState(env: domain.Environment) {
  if (!state.value) return
  state.value = new domain.WorkspaceState({
    ...state.value,
    environments: (state.value.environments ?? []).map((item) => item.id === env.id ? new domain.Environment({ ...item, ...env }) : item)
  })
}

function updateGlobalsInState(globals: domain.KeyValue[]) {
  if (!state.value) return
  state.value = new domain.WorkspaceState({
    ...state.value,
    globals: cloneKeyValues(globals)
  })
}

function scheduleRequestAutosave() {
  if (suppressRequestAutosave || !activeRequest.value?.id) return
  const snapshot = requestSnapshot(activeRequest.value)
  if (!snapshot || (snapshot === lastRequestAutosaveSnapshot && activeRequest.value.id === lastRequestAutosaveId)) return
  clearRequestAutosaveTimer()
  requestAutosaveTimer = setTimeout(() => { void runRequestAutosave() }, 650)
}

function scheduleEnvironmentAutosave() {
  if (suppressEnvironmentAutosave || !envDraft.id) return
  const snapshot = environmentDraftSnapshot()
  if (!snapshot || snapshot === lastEnvironmentAutosaveSnapshot) return
  clearEnvironmentAutosaveTimer()
  environmentAutosaveTimer = setTimeout(() => { void runEnvironmentAutosave() }, 650)
}

function scheduleGlobalsAutosave() {
  if (suppressGlobalsAutosave) return
  const snapshot = globalsSnapshot()
  if (snapshot === lastGlobalsAutosaveSnapshot) return
  clearGlobalsAutosaveTimer()
  globalsAutosaveTimer = setTimeout(() => { void runGlobalsAutosave() }, 650)
}

function scheduleSettingsAutosave() {
  if (suppressSettingsAutosave) return
  const snapshot = settingsSnapshot()
  if (snapshot === lastSettingsAutosaveSnapshot) return
  clearSettingsAutosaveTimer()
  settingsAutosaveTimer = setTimeout(() => { void runSettingsAutosave() }, 650)
}

async function runRequestAutosave() {
  clearRequestAutosaveTimer()
  if (requestAutosavePromise) {
    requestAutosaveQueued = true
    return requestAutosavePromise
  }
  requestAutosavePromise = (async () => {
    try {
      do {
        requestAutosaveQueued = false
        await saveActiveRequestAutosaveOnce()
      } while (requestAutosaveQueued || (requestSnapshot(activeRequest.value) && requestSnapshot(activeRequest.value) !== lastRequestAutosaveSnapshot))
    } finally {
      requestAutosavePromise = null
    }
  })()
  return requestAutosavePromise
}

async function runEnvironmentAutosave() {
  clearEnvironmentAutosaveTimer()
  if (environmentAutosavePromise) {
    environmentAutosaveQueued = true
    return environmentAutosavePromise
  }
  environmentAutosavePromise = (async () => {
    try {
      do {
        environmentAutosaveQueued = false
        await saveEnvironmentAutosaveOnce()
      } while (environmentAutosaveQueued || (environmentDraftSnapshot() && environmentDraftSnapshot() !== lastEnvironmentAutosaveSnapshot))
    } finally {
      environmentAutosavePromise = null
    }
  })()
  return environmentAutosavePromise
}

async function runGlobalsAutosave() {
  clearGlobalsAutosaveTimer()
  if (globalsAutosavePromise) {
    globalsAutosaveQueued = true
    return globalsAutosavePromise
  }
  globalsAutosavePromise = (async () => {
    try {
      do {
        globalsAutosaveQueued = false
        await saveGlobalsAutosaveOnce()
      } while (globalsAutosaveQueued || globalsSnapshot() !== lastGlobalsAutosaveSnapshot)
    } finally {
      globalsAutosavePromise = null
    }
  })()
  return globalsAutosavePromise
}

async function runSettingsAutosave() {
  clearSettingsAutosaveTimer()
  if (settingsAutosavePromise) {
    settingsAutosaveQueued = true
    return settingsAutosavePromise
  }
  settingsAutosavePromise = (async () => {
    try {
      do {
        settingsAutosaveQueued = false
        await saveSettingsAutosaveOnce()
      } while (settingsAutosaveQueued || settingsSnapshot() !== lastSettingsAutosaveSnapshot)
    } finally {
      settingsAutosavePromise = null
    }
  })()
  return settingsAutosavePromise
}

async function flushRequestAutosave() {
  clearRequestAutosaveTimer()
  await runRequestAutosave()
}

async function flushEnvironmentAutosave() {
  clearEnvironmentAutosaveTimer()
  await runEnvironmentAutosave()
}

async function flushGlobalsAutosave() {
  clearGlobalsAutosaveTimer()
  await runGlobalsAutosave()
}

async function flushSettingsAutosave() {
  clearSettingsAutosaveTimer()
  await runSettingsAutosave()
}

async function flushCurrentEnvironmentPanelAutosave() {
  if (environmentPanel.value === 'environment') {
    await flushEnvironmentAutosave()
  } else {
    await flushGlobalsAutosave()
  }
}

async function saveActiveRequestAutosaveOnce() {
  const request = activeRequest.value
  if (!request?.id) return
  const nextRequest = normalizeRequest(cloneRequest(request))
  syncFormBody(nextRequest)
  const snapshot = requestSnapshot(nextRequest)
  if (!snapshot || (snapshot === lastRequestAutosaveSnapshot && nextRequest.id === lastRequestAutosaveId)) return
  try {
    await SaveRequest(nextRequest)
    lastRequestAutosaveId = nextRequest.id
    lastRequestAutosaveSnapshot = snapshot
    if (activeRequest.value?.id === nextRequest.id) {
      updateRequestInState(nextRequest)
    }
    statusMessage.value = t.value.requestSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function saveEnvironmentAutosaveOnce() {
  if (!envDraft.id) return
  const env = environmentDraftModel()
  const snapshot = environmentDraftSnapshot()
  if (!snapshot || snapshot === lastEnvironmentAutosaveSnapshot) return
  try {
    await SaveEnvironment(env)
    lastEnvironmentAutosaveSnapshot = snapshot
    updateEnvironmentInState(env)
    statusMessage.value = t.value.environmentSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function saveGlobalsAutosaveOnce() {
  const globals = normalizedKeyValues(globalsDraft.value)
  const snapshot = JSON.stringify(globals)
  if (snapshot === lastGlobalsAutosaveSnapshot) return
  try {
    await SaveGlobals(globals)
    lastGlobalsAutosaveSnapshot = snapshot
    updateGlobalsInState(globals)
    statusMessage.value = t.value.globalsSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function saveSettingsAutosaveOnce() {
  const nextSettings = settingsModel()
  const snapshot = JSON.stringify(nextSettings)
  if (snapshot === lastSettingsAutosaveSnapshot) return
  try {
    const next = await SaveSettings(nextSettings)
    lastSettingsAutosaveSnapshot = snapshot
    setState(next)
    statusMessage.value = t.value.settingsSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function loadState() {
  try {
    const next = await GetState()
    setState(next)
    statusMessage.value = t.value.workspaceLoaded
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

function setState(next: domain.WorkspaceState) {
  state.value = next
  const collections = next.collections ?? []
  const activeCollectionExists = collections.some((collection) => collection.id === activeCollectionId.value)
  if (!activeCollectionId.value || !activeCollectionExists) {
    activeCollectionId.value = collections[0]?.id ?? ''
  }
  const collection = collections.find((item) => item.id === activeCollectionId.value)
  if (!activeRequest.value) {
    const first = collection?.requests?.[0]
    if (first) activeRequest.value = normalizeRequest(cloneRequest(first))
  } else {
    const fresh = collection?.requests?.find((request) => request.id === activeRequest.value?.id)
    if (fresh) {
      activeRequest.value = normalizeRequest(cloneRequest(fresh))
    } else {
      activeRequest.value = collection?.requests?.[0] ? normalizeRequest(cloneRequest(collection.requests[0])) : null
      response.value = null
    }
  }
  ensureRunnerRequest(activeRequest.value?.id ?? '')
}

async function selectRequest(request: domain.Request) {
  await flushRequestAutosave()
  activeRequest.value = normalizeRequest(cloneRequest(request))
  setRunnerRequest(request.id)
  response.value = null
  activeResponseTab.value = 'body'
}

async function selectHistoryRequest(request: domain.Request) {
  await selectRequest(request)
  activeNav.value = 'collections'
}

async function createCollection() {
  await flushRequestAutosave()
  const name = `${t.value.collections} ${(state.value?.collections?.length ?? 0) + 1}`
  const next = await CreateCollection(name)
  setState(next)
  activeCollectionId.value = next.collections[next.collections.length - 1]?.id ?? activeCollectionId.value
  statusMessage.value = t.value.collectionCreated
}

async function selectCollection(collection: domain.Collection) {
  await flushRequestAutosave()
  activeCollectionId.value = collection.id
  activeRequest.value = collection.requests?.[0] ? normalizeRequest(cloneRequest(collection.requests[0])) : null
  setRunnerRequest(activeRequest.value?.id ?? '')
  response.value = null
  collectionPickerOpen.value = false
  pendingDeleteCollectionId.value = ''
}

function selectCollectionById(id: string) {
  const collection = state.value?.collections?.find((item) => item.id === id)
  if (collection) void selectCollection(collection)
}

function startEditingCollection(collection: domain.Collection) {
  editingCollectionId.value = collection.id
  editingCollectionName.value = collection.name
  pendingDeleteCollectionId.value = ''
}

function cancelEditingCollection() {
  editingCollectionId.value = ''
  editingCollectionName.value = ''
}

async function saveEditingCollection(collection: domain.Collection) {
  if (editingCollectionId.value !== collection.id) return
  const nextName = editingCollectionName.value.trim()
  cancelEditingCollection()
  if (!nextName || nextName === collection.name) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await SaveCollection(new domain.Collection({
      ...collection,
      name: nextName
    }))
    setState(next)
    activeCollectionId.value = collection.id
    statusMessage.value = t.value.collectionSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function deleteCollectionFromPicker(collection: domain.Collection) {
  if (!collection.id) return
  const hasItems = (collection.requests?.length ?? 0) > 0 || (collection.folders?.length ?? 0) > 0
  if (hasItems && pendingDeleteCollectionId.value !== collection.id) {
    pendingDeleteCollectionId.value = collection.id
    return
  }

  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await DeleteCollection(collection.id)
    if (activeCollectionId.value === collection.id) {
      activeCollectionId.value = next.collections?.[0]?.id ?? ''
      activeRequest.value = next.collections?.[0]?.requests?.[0] ? normalizeRequest(cloneRequest(next.collections[0].requests[0])) : null
      response.value = null
    }
    setState(next)
    pendingDeleteCollectionId.value = ''
    editingCollectionId.value = ''
    statusMessage.value = t.value.collectionDeleted
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

function makeNewRequest(collection: domain.Collection) {
  return new domain.Request({
    id: crypto.randomUUID(),
    collectionId: collection.id,
    parentId: '',
    name: t.value.requestName,
    method: 'GET',
    url: '{{baseUrl}}/anything',
    params: [newKeyValue()],
    headers: [new domain.KeyValue({ id: crypto.randomUUID(), enabled: true, key: 'Accept', value: 'application/json', description: '', secret: false })],
    bodyMode: 'none',
    body: '',
    formItems: [newFormItem()],
    auth: new domain.AuthConfig({ type: 'none', values: {} }),
    proxy: new domain.ProxyConfig({ mode: 'inherit', url: '', noProxy: '' }),
    preScript: '',
    testScript: 'pm.test("Status is successful", function () { expect(pm.response.code).to.be.ok(); });',
    timeoutMs: 30000,
    sortOrder: collection.requests?.length ?? 0
  })
}

async function createRequest() {
  await flushRequestAutosave()
  busy.value = true
  try {
    let collection = activeCollection.value
    if (!collection) {
      const refreshedState = await GetState()
      setState(refreshedState)
      collection = activeCollection.value
    }
    if (!collection) {
      const createdState = await CreateCollection(`${t.value.collections} ${(state.value?.collections?.length ?? 0) + 1}`)
      collection = createdState.collections?.[createdState.collections.length - 1] ?? createdState.collections?.[0] ?? null
      setState(createdState)
      if (collection?.id) activeCollectionId.value = collection.id
    }
    if (!collection?.id) return

    const request = makeNewRequest(collection)
    const next = await SaveRequest(request)
    activeCollectionId.value = collection.id
    setState(next)
    const saved = next.collections
      ?.find((item) => item.id === collection.id)
      ?.requests
      ?.find((item) => item.id === request.id)
    await selectRequest(saved ?? request)
    activeRequestTab.value = 'params'
    statusMessage.value = t.value.requestCreated
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function sendActiveRequest() {
  if (!activeRequest.value) return
  await flushRequestAutosave()
  syncFormBody(activeRequest.value)
  const requestToSend = normalizeRequest(cloneRequest(activeRequest.value))
  if (!requestToSend.id) requestToSend.id = crypto.randomUUID()
  activeRequest.value = normalizeRequest(cloneRequest(requestToSend))
  busy.value = true
  response.value = null
  try {
    statusMessage.value = t.value.sendingRequest
    const savedState = await SaveRequest(requestToSend)
    setState(savedState)
    const result = await SendRequest(requestToSend, activeEnvironment.value?.id ?? '', globalsDraft.value)
    response.value = result
    requestPreview.value = result.request || requestPreview.value
    if (result.contentType && result.body) {
      result.body = await FormatBody(result.contentType, result.body)
    }
    const latestState = await GetState()
    setState(latestState)
    statusMessage.value = result.error ? result.error : `${result.statusCode} in ${result.durationMs} ms`
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function refreshRequestPreview() {
  if (!activeRequest.value) {
    requestPreview.value = null
    return
  }
  requestPreviewBusy.value = true
  try {
    const requestToPreview = normalizeRequest(cloneRequest(activeRequest.value))
    syncFormBody(requestToPreview)
    requestPreview.value = await PreviewRequest(requestToPreview, activeEnvironment.value?.id ?? '', globalsDraft.value)
  } catch (error) {
    requestPreview.value = new domain.PreparedRequest({ error: formatError(error) })
  } finally {
    requestPreviewBusy.value = false
  }
}

async function refreshVariableDebug() {
  variableDebugBusy.value = true
  try {
    const requestToDebug = activeRequest.value ? normalizeRequest(cloneRequest(activeRequest.value)) : new domain.Request()
    if (activeRequest.value) syncFormBody(requestToDebug)
    variableDebugReport.value = await DebugRequestVariables(requestToDebug, activeEnvironment.value?.id ?? '', globalsDraft.value)
  } catch (error) {
    variableDebugReport.value = new domain.VariableDebugReport({ variables: [], errors: [formatError(error)] })
  } finally {
    variableDebugBusy.value = false
  }
}

async function testEnvironmentJSONPathVariable(variable: domain.KeyValue) {
  if (!envDraft.id) throw new Error(t.value.environmentName)
  const variableToTest = new domain.KeyValue({
    ...variable,
    key: variable.key?.trim() || `__jsonpath_${variable.id || crypto.randomUUID()}`,
    enabled: true,
    valueType: 'responseJsonPath'
  })
  return TestVariable(variableToTest, envDraft.id, globalsDraft.value)
}

async function importPostmanCollection() {
  if (!postmanText.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await ImportPostmanCollection(postmanText.value)
    setState(next)
    activeModal.value = null
    postmanText.value = ''
    statusMessage.value = t.value.collectionImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function importOpenAPICollection() {
  if (!openAPIText.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = selectedOpenAPIServer.value
      ? await ImportOpenAPICollectionWithOptions(openAPIText.value, new domain.OpenAPIImportOptions({ serverUrl: selectedOpenAPIServer.value }))
      : await ImportOpenAPICollection(openAPIText.value)
    setState(next)
    activeModal.value = null
    openAPIText.value = ''
    openAPIServers.value = []
    selectedOpenAPIServer.value = ''
    statusMessage.value = t.value.openAPIImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function importSwaggerCollection() {
  if (!swaggerUrl.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await ImportSwaggerURL(swaggerUrl.value)
    setState(next)
    activeModal.value = null
    swaggerUrl.value = ''
    statusMessage.value = t.value.swaggerImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function inspectOpenAPIText() {
  if (!openAPIText.value.trim()) {
    openAPIServers.value = []
    selectedOpenAPIServer.value = ''
    return
  }
  try {
    const info = await InspectOpenAPI(openAPIText.value)
    openAPIServers.value = info.servers ?? []
    if (!selectedOpenAPIServer.value) selectedOpenAPIServer.value = openAPIServers.value[0] ?? ''
  } catch {
    openAPIServers.value = []
    selectedOpenAPIServer.value = ''
  }
}

async function importHARCollection() {
  if (!harText.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await ImportHARCollection(harText.value)
    setState(next)
    activeModal.value = null
    harText.value = ''
    statusMessage.value = t.value.harImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function importFetchRequest() {
  if (!fetchText.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const previousRequestIds = new Set(activeCollection.value?.requests?.map((request) => request.id) ?? [])
    const collectionID = activeCollection.value?.id ?? ''
    const next = await ImportFetchRequest(fetchText.value, collectionID)
    setState(next)
    await selectLatestImportedRequest(previousRequestIds)
    activeRequestTab.value = 'headers'
    activeModal.value = null
    fetchText.value = ''
    statusMessage.value = t.value.fetchImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function importCurlRequest() {
  if (!curlText.value.trim()) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const previousRequestIds = new Set(activeCollection.value?.requests?.map((request) => request.id) ?? [])
    const collectionID = activeCollection.value?.id ?? ''
    const next = await ImportCurlRequest(curlText.value, collectionID)
    setState(next)
    await selectLatestImportedRequest(previousRequestIds)
    activeRequestTab.value = 'headers'
    activeModal.value = null
    curlText.value = ''
    statusMessage.value = t.value.curlImported
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function selectLatestImportedRequest(previousRequestIds: Set<string>) {
  const imported = activeCollection.value?.requests?.find((request) => !previousRequestIds.has(request.id))
  if (imported) await selectRequest(imported)
}

function openPostmanModal() {
  closeCollectionMenus()
  postmanText.value = ''
  activeModal.value = 'postman'
}

function openOpenAPIModal() {
  closeCollectionMenus()
  openAPIText.value = ''
  openAPIServers.value = []
  selectedOpenAPIServer.value = ''
  activeModal.value = 'openapi'
}

function openSwaggerModal() {
  closeCollectionMenus()
  swaggerUrl.value = ''
  openAPIServers.value = []
  selectedOpenAPIServer.value = ''
  activeModal.value = 'swagger'
}

function openHARModal() {
  closeCollectionMenus()
  harText.value = ''
  activeModal.value = 'har'
}

function openFetchModal() {
  closeCollectionMenus()
  fetchText.value = ''
  activeModal.value = 'fetch'
}

function openCurlModal() {
  closeCollectionMenus()
  curlText.value = ''
  activeModal.value = 'curl'
}

function closeModal() {
  activeModal.value = null
}

function submitActiveModal() {
  if (activeModal.value === 'postman') return importPostmanCollection()
  if (activeModal.value === 'openapi') return importOpenAPICollection()
  if (activeModal.value === 'swagger') return importSwaggerCollection()
  if (activeModal.value === 'har') return importHARCollection()
  if (activeModal.value === 'fetch') return importFetchRequest()
  if (activeModal.value === 'curl') return importCurlRequest()
}

async function importActiveModalFromFile() {
  if (!activeModal.value || activeModal.value === 'export') return
  try {
    const content = await OpenTextFile(activeModalTitle.value)
    if (!content) return
    if (activeModal.value === 'postman') postmanText.value = content
    if (activeModal.value === 'openapi') {
      openAPIText.value = content
      await inspectOpenAPIText()
    }
    if (activeModal.value === 'har') harText.value = content
    if (activeModal.value === 'fetch') fetchText.value = content
    if (activeModal.value === 'curl') curlText.value = content
    statusMessage.value = t.value.fileLoaded
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportActiveModalToFile() {
  if (activeModal.value !== 'export' || !exportText.value) return
  try {
    const path = await SaveTextFile(t.value.exportToFile, exportFilename.value, exportText.value)
    if (path) statusMessage.value = t.value.fileSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportCollection() {
  const collection = activeCollection.value
  if (!collection) return
  await flushRequestAutosave()
  try {
    exportText.value = await ExportPostmanCollection(collection.id)
    exportFilename.value = `${safeFilename(collection.name || 'restdeck-collection')}.postman_collection.json`
    activeModal.value = 'export'
    closeCollectionMenus()
    statusMessage.value = t.value.collectionExported
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportOpenAPICollection() {
  const collection = activeCollection.value
  if (!collection) return
  await flushRequestAutosave()
  try {
    exportText.value = await ExportOpenAPICollection(collection.id)
    exportFilename.value = `${safeFilename(collection.name || 'restdeck-openapi')}.openapi.json`
    activeModal.value = 'export'
    closeCollectionMenus()
    statusMessage.value = t.value.openAPIExported
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportHARCollection() {
  const collection = activeCollection.value
  if (!collection) return
  await flushRequestAutosave()
  try {
    exportText.value = await ExportHARCollection(collection.id)
    exportFilename.value = `${safeFilename(collection.name || 'restdeck')}.har`
    activeModal.value = 'export'
    closeCollectionMenus()
    statusMessage.value = t.value.harExported
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportRequestFromMenu(request: domain.Request) {
  const source = requestActionSource(request)
  const collection = collectionForRequest(source)
  try {
    exportText.value = await ExportPostmanRequest(source, collection?.name ?? t.value.collections)
    exportFilename.value = `${safeFilename(source.name || 'restdeck-request')}.postman_request.json`
    activeModal.value = 'export'
    statusMessage.value = t.value.requestExported
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

function closeCollectionMenus() {
  collectionPickerOpen.value = false
  optionsMenuOpen.value = false
}

function requestActionSource(request: domain.Request) {
  const source = activeRequest.value?.id === request.id ? activeRequest.value : request
  const next = normalizeRequest(cloneRequest(source))
  syncFormBody(next)
  return next
}

function collectionForRequest(request: domain.Request) {
  return state.value?.collections?.find((collection) => collection.id === request.collectionId)
    ?? activeCollection.value
    ?? null
}

function safeFilename(value: string) {
  return value.trim().replace(/[<>:"/\\|?*\x00-\x1F]/g, '_').replace(/\s+/g, ' ').slice(0, 80) || 'restdeck'
}

async function saveOrderedRequests(collectionID: string, requests: domain.Request[]) {
  let nextState: domain.WorkspaceState | null = null
  for (const [index, request] of requests.entries()) {
    const nextRequest = normalizeRequest(cloneRequest(request))
    nextRequest.collectionId = collectionID
    nextRequest.sortOrder = index
    syncFormBody(nextRequest)
    nextState = await SaveRequest(nextRequest)
  }
  if (nextState) {
    activeCollectionId.value = collectionID
    setState(nextState)
  }
  return nextState
}

function openRequestCodeModal(request: domain.Request) {
  codeModalRequest.value = requestActionSource(request)
}

function setCodeModalVisible(value: boolean) {
  if (!value) codeModalRequest.value = null
}

function markCodeCopied() {
  statusMessage.value = t.value.codeCopied
}

async function copyRunnerReport() {
  try {
    await ClipboardSetText(runnerReportText())
    statusMessage.value = t.value.runnerReportCopied
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function exportRunnerReport() {
  try {
    const path = await SaveTextFile(t.value.exportReport, 'restdeck-runner-report.md', runnerReportText())
    if (path) statusMessage.value = path
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function deleteCookie(cookie: domain.Cookie) {
  try {
    const next = await DeleteCookie(cookie)
    setState(next)
    statusMessage.value = t.value.cookieDeleted
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function clearCookies() {
  try {
    const next = await ClearCookies()
    setState(next)
    statusMessage.value = t.value.cookiesCleared
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function pinRequestToTop(request: domain.Request) {
  await flushRequestAutosave()
  const collection = collectionForRequest(request)
  if (!collection?.id || !request.id) return
  busy.value = true
  try {
    const source = requestActionSource(request)
    const requests = (collection.requests ?? []).map((item) => item.id === source.id ? source : item)
    const ordered = [source, ...requests.filter((item) => item.id !== source.id)]
    await saveOrderedRequests(collection.id, ordered)
    statusMessage.value = t.value.requestPinned
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function duplicateRequestFromMenu(request: domain.Request) {
  await flushRequestAutosave()
  const collection = collectionForRequest(request)
  if (!collection?.id || !request.id) return
  busy.value = true
  try {
    const source = requestActionSource(request)
    const duplicate = normalizeRequest(cloneRequest(source))
    duplicate.id = crypto.randomUUID()
    duplicate.collectionId = collection.id
    duplicate.name = `${source.name || t.value.requestName} ${t.value.copySuffix}`
    duplicate.sortOrder = (source.sortOrder ?? 0) + 1
    syncFormBody(duplicate)

    const ordered: domain.Request[] = []
    const existing = (collection.requests ?? []).map((item) => item.id === source.id ? source : item)
    for (const item of existing) {
      ordered.push(item)
      if (item.id === source.id) ordered.push(duplicate)
    }
    if (!ordered.some((item) => item.id === duplicate.id)) ordered.push(duplicate)

    const next = await saveOrderedRequests(collection.id, ordered)
    const created = next?.collections
      ?.find((item) => item.id === collection.id)
      ?.requests
      ?.find((item) => item.id === duplicate.id)
    if (created) await selectRequest(created)
    statusMessage.value = t.value.requestCopied
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function deleteRequestFromMenu(request: domain.Request) {
  if (!request.id) return
  await flushRequestAutosave()
  busy.value = true
  try {
    const next = await DeleteRequest(request.id)
    if (activeRequest.value?.id === request.id) {
      activeRequest.value = null
      response.value = null
    }
    setState(next)
    statusMessage.value = t.value.requestDeleted
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function createEnvironment() {
  await flushCurrentEnvironmentPanelAutosave()
  busy.value = true
  try {
    const next = await CreateEnvironment(`${t.value.environments} ${(state.value?.environments?.length ?? 0) + 1}`)
    setState(next)
    environmentPanel.value = 'environment'
    statusMessage.value = t.value.environmentCreated
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function deleteEnvironment(id: string) {
  if (!id) return
  await flushCurrentEnvironmentPanelAutosave()
  busy.value = true
  try {
    const next = await DeleteEnvironment(id)
    setState(next)
    statusMessage.value = t.value.environmentDeleted
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function renameEnvironment(env: domain.Environment, name: string) {
  const nextName = name.trim()
  if (!env?.id || !nextName || nextName === env.name) return
  await flushCurrentEnvironmentPanelAutosave()
  busy.value = true
  try {
    const variables = env.id === envDraft.id ? envDraft.variables.map(normalizeKeyValue) : cloneKeyValues(env.variables ?? [])
    const next = await SaveEnvironment(new domain.Environment({
      ...env,
      name: nextName,
      variables,
      isActive: env.id === activeEnvironment.value?.id || env.isActive
    }))
    setState(next)
    statusMessage.value = t.value.environmentSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function setEnvironment(id: string) {
  await flushRequestAutosave()
  await flushCurrentEnvironmentPanelAutosave()
  await flushSettingsAutosave()
  environmentPanel.value = 'environment'
  const next = await SetActiveEnvironment(id)
  setState(next)
  statusMessage.value = t.value.environmentSelected
}

async function selectGlobalEnvironment() {
  await flushCurrentEnvironmentPanelAutosave()
  environmentPanel.value = 'globals'
}

async function runWebSocketCheck() {
  realtimeBusy.value = true
  wsResult.value = null
  try {
    wsResult.value = await TestWebSocket(new realtime.WebSocketRequest(wsDraft), activeEnvironment.value?.id ?? '', globalsDraft.value)
    statusMessage.value = wsResult.value.error ? wsResult.value.error : `WebSocket ${t.value.received} ${wsResult.value.received?.length ?? 0}`
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    realtimeBusy.value = false
  }
}

async function runSSECheck() {
  realtimeBusy.value = true
  sseResult.value = null
  try {
    sseResult.value = await TestSSE(new realtime.SSERequest(sseDraft), activeEnvironment.value?.id ?? '', globalsDraft.value)
    statusMessage.value = sseResult.value.error ? sseResult.value.error : `SSE ${t.value.received} ${sseResult.value.events?.length ?? 0}`
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    realtimeBusy.value = false
  }
}

function addParam() {
  activeRequest.value?.params.push(newKeyValue())
}

function addHeader() {
  activeRequest.value?.headers.push(newKeyValue())
}

function addFormItem() {
  const request = activeRequest.value
  if (!request) return
  request.formItems = request.formItems ?? []
  request.formItems.push(newFormItem())
  syncFormBody(request)
}

function addVariable(target: domain.KeyValue[]) {
  target.push(newKeyValue())
}

function removeRow(target: domain.KeyValue[], index: number) {
  target.splice(index, 1)
}

function removeFormItem(index: number) {
  const request = activeRequest.value
  if (!request) return
  request.formItems.splice(index, 1)
  syncFormBody(request)
}

async function selectFormFile(index: number) {
  const request = activeRequest.value
  const item = request?.formItems?.[index]
  if (!request || !item) return
  try {
    const path = await SelectFile()
    if (!path) return
    item.filePath = path
    item.type = 'file'
    syncFormBody(request)
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

function setTheme(value: Theme) {
  theme.value = value
}

async function runResponseJSONPathQuery() {
  if (!response.value?.body || !responseJSONPath.value.trim()) return
  try {
    responseJSONPathResult.value = await QueryJSONPath(response.value.body, responseJSONPath.value.trim())
  } catch (error) {
    responseJSONPathResult.value = ''
    statusMessage.value = formatError(error)
  }
}

async function copyResponseValue(value: string) {
  try {
    await ClipboardSetText(value)
    statusMessage.value = t.value.jsonPathCopied
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function saveResponseBody() {
  if (!response.value) return
  try {
    const path = await SaveTextFile(t.value.saveResponse, 'restdeck-response.txt', response.value.body ?? '')
    if (path) statusMessage.value = t.value.responseSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

async function createResponseVariable() {
  if (!activeRequest.value?.id || !responseJSONPath.value.trim()) return
  try {
    const key = responseVariableKey.value.trim() || responseJSONPath.value.trim().split(/[.[\]]/).filter(Boolean).pop() || 'responseValue'
    const next = await CreateResponseVariable(activeEnvironment.value?.id ?? '', key, activeRequest.value.id, responseJSONPath.value.trim(), responseJSONPathResult.value)
    setState(next)
    const updatedEnv = activeEnvironment.value
    if (updatedEnv) {
      envDraft.id = updatedEnv.id
      envDraft.name = updatedEnv.name
      envDraft.variables = cloneKeyValues(updatedEnv.variables ?? [])
      lastEnvironmentAutosaveSnapshot = environmentDraftSnapshot()
    }
    responseVariableKey.value = ''
    statusMessage.value = t.value.responseVariableCreated
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

function setAuthType(value: string) {
  if (!activeRequest.value) return
  activeRequest.value.auth = new domain.AuthConfig({ type: value, values: defaultAuthValues(value) })
}

function minimiseWindow() {
  WindowMinimise()
}

function toggleWindowMaximise() {
  WindowToggleMaximise()
}

async function closeWindow() {
  await flushRequestAutosave()
  await flushCurrentEnvironmentPanelAutosave()
  await flushSettingsAutosave()
  Quit()
}


</script>

<template>
  <div class="app-shell">
    <AppTitleBar
      v-model:search="search"
      :t="t"
      :theme="theme"
      :active-environment="activeEnvironment"
      :environments="state?.environments ?? []"
      @toggle-theme="setTheme(theme === 'dark' ? 'light' : 'dark')"
      @select-environment="setEnvironment"
      @minimize="minimiseWindow"
      @toggle-maximize="toggleWindowMaximise"
      @close="closeWindow"
    />

    <main :class="['workspace', { 'workspace-no-sidebar': activeNav === 'history' || activeNav === 'runner' || activeNav === 'settings' }]">
      <SidebarRail v-model:active-nav="activeNav" :items="navItems" />

      <WorkspaceSidebar
        v-if="activeNav !== 'history' && activeNav !== 'runner' && activeNav !== 'settings'"
        v-model:collection-picker-open="collectionPickerOpen"
        v-model:options-menu-open="optionsMenuOpen"
        v-model:editing-collection-name="editingCollectionName"
        :t="t"
        :active-nav="activeNav"
        :nav-label="navItems.find((item) => item.key === activeNav)?.label ?? ''"
        :collections="state?.collections ?? []"
        :active-collection="activeCollection"
        :filtered-requests="filteredRequests"
        :active-request="activeRequest"
        :environments="state?.environments ?? []"
        :active-environment="activeEnvironment"
        :environment-panel="environmentPanel"
        :editing-collection-id="editingCollectionId"
        :pending-delete-collection-id="pendingDeleteCollectionId"
        @select-collection="selectCollection"
        @start-editing-collection="startEditingCollection"
        @cancel-editing-collection="cancelEditingCollection"
        @save-editing-collection="saveEditingCollection"
        @delete-collection="deleteCollectionFromPicker"
        @create-collection="createCollection"
        @create-request="createRequest"
        @open-fetch-modal="openFetchModal"
        @open-curl-modal="openCurlModal"
        @open-postman-modal="openPostmanModal"
        @open-open-a-p-i-modal="openOpenAPIModal"
        @open-swagger-modal="openSwaggerModal"
        @open-h-a-r-modal="openHARModal"
        @export-collection="exportCollection"
        @export-open-a-p-i-collection="exportOpenAPICollection"
        @export-h-a-r-collection="exportHARCollection"
        @select-request="selectRequest"
        @generate-request-code="openRequestCodeModal"
        @export-request="exportRequestFromMenu"
        @pin-request="pinRequestToTop"
        @duplicate-request="duplicateRequestFromMenu"
        @delete-request="deleteRequestFromMenu"
        @create-environment="createEnvironment"
        @select-environment="setEnvironment"
        @select-global-environment="selectGlobalEnvironment"
        @rename-environment="renameEnvironment"
        @delete-environment="deleteEnvironment"
      />

      <section :class="['main-pane', {
        'runner-main-pane': activeNav === 'runner',
        'settings-main-pane': activeNav === 'settings',
        'realtime-main-pane': activeNav === 'realtime'
      }]">
        <RequestWorkspace
          v-if="activeNav === 'collections'"
          v-model:active-request="activeRequest"
          v-model:active-request-tab="activeRequestTab"
          v-model:active-response-tab="activeResponseTab"
          v-model:response-view="responseView"
          :t="t"
          :active-collection="activeCollection"
          :response="response"
          :request-tabs="requestTabs"
          :response-tabs="responseTabs"
          :highlighted-response-body="highlightedResponseBody"
          :pretty-response-body="prettyResponseBody"
          v-model:response-search="responseSearch"
          v-model:responseJSONPath="responseJSONPath"
          v-model:response-variable-key="responseVariableKey"
          :response-search-matches="responseSearchMatches"
          :responseJSONPathResult="responseJSONPathResult"
          :responseJSONPathOptions="responseJSONPathOptions"
          :busy="busy"
          :methods="methods"
          :auth-types="authTypes"
          :body-modes="bodyModes"
          :variable-suggestions="variableSuggestions"
          :request-preview="requestPreview"
          :request-preview-busy="requestPreviewBusy"
          :variable-debug-report="variableDebugReport"
          :variable-debug-busy="variableDebugBusy"
          :send-request-action="sendActiveRequest"
          @create-request="createRequest"
          @add-param="addParam"
          @add-header="addHeader"
          @add-form-item="addFormItem"
          @remove-row="removeRow"
          @remove-form-item="removeFormItem"
          @select-form-file="selectFormFile"
          @set-auth-type="setAuthType"
          @query-json-path="runResponseJSONPathQuery"
          @copy-response-value="copyResponseValue"
          @save-response="saveResponseBody"
          @create-response-variable="createResponseVariable"
          @refresh-request-preview="refreshRequestPreview"
          @refresh-variable-debug="refreshVariableDebug"
        />

        <EnvironmentsView
          v-else-if="activeNav === 'environments'"
          v-model:env-draft="envDraft"
          v-model:globals-draft="globalsDraft"
          :t="t"
          :mode="environmentPanel"
          :collections="state?.collections ?? []"
          :variable-suggestions="variableSuggestions"
          :test-json-path="testEnvironmentJSONPathVariable"
          @add-variable="addVariable"
          @remove-row="removeRow"
        />

        <HistoryView
          v-else-if="activeNav === 'history'"
          :t="t"
          :history="state?.history ?? []"
          @select-request="selectHistoryRequest"
        />

        <RunnerView
          v-else-if="activeNav === 'runner'"
          :t="t"
          :collections="state?.collections ?? []"
          :active-collection-id="activeCollectionId"
          :active-request-id="runnerRequestId"
          :runner-scope="runnerScope"
          :runner-iterations="runnerIterations"
          :active-environment="activeEnvironment"
          :active-collection="activeCollection"
          :runner-result="runnerResult"
          :runner-queue="runnerQueue"
          :runner-busy="runnerBusy"
          :runner-failure-policy="runnerFailurePolicy"
          :runner-retry-count="runnerRetryCount"
          :runner-delay-ms="runnerDelayMs"
          :runner-history="state?.runnerHistory ?? []"
          @select-collection="selectCollectionById"
          @select-request="selectRunnerRequest"
          @set-scope="setRunnerScope"
          @set-iterations="setRunnerIterations"
          @set-failure-policy="setRunnerFailurePolicy"
          @set-retry-count="setRunnerRetryCount"
          @set-delay-ms="setRunnerDelayMs"
          @run-collection="runActiveCollection"
          @run-request="runRunnerRequest"
          @stop-run="stopRunner"
          @copy-report="copyRunnerReport"
          @export-report="exportRunnerReport"
        />

        <RealtimeView
          v-else-if="activeNav === 'realtime'"
          v-model:ws-draft="wsDraft"
          v-model:sse-draft="sseDraft"
          :t="t"
          :variable-suggestions="variableSuggestions"
          :realtime-busy="realtimeBusy"
          :ws-result="wsResult"
          :sse-result="sseResult"
          @run-web-socket="runWebSocketCheck"
          @run-s-s-e="runSSECheck"
        />

        <SettingsView
          v-else
          v-model:language="language"
          v-model:settings-draft="settingsDraft"
          :t="t"
          :cookies="state?.cookies ?? []"
          @delete-cookie="deleteCookie"
          @clear-cookies="clearCookies"
        />
      </section>
    </main>

    <footer class="statusbar">
      <span>{{ statusMessage }}</span>
    </footer>
  </div>

  <ImportModal
    v-model:postman-text="postmanText"
    v-model:open-a-p-i-text="openAPIText"
    v-model:swagger-url="swaggerUrl"
    v-model:har-text="harText"
    v-model:selected-open-a-p-i-server="selectedOpenAPIServer"
    v-model:fetch-text="fetchText"
    v-model:curl-text="curlText"
    :active-modal="activeModal"
    :title="activeModalTitle"
    :busy="busy"
    :t="t"
    :export-text="exportText"
    :open-a-p-i-servers="openAPIServers"
    @close="closeModal"
    @submit="submitActiveModal"
    @import-from-file="importActiveModalFromFile"
    @export-to-file="exportActiveModalToFile"
  />

  <RequestCodeModal
    :visible="!!codeModalRequest"
    :request="codeModalRequest"
    :t="t"
    @update:visible="setCodeModalVisible"
    @copied="markCodeCopied"
  />
</template>
