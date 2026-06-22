<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import {
  Globe2,
  History,
  ListTree,
  Play,
  Settings,
  Radio
} from 'lucide-vue-next'
import { Quit, WindowMinimise, WindowToggleMaximise } from '../wailsjs/runtime/runtime'
import {
  CreateCollection,
  DeleteCollection,
  DeleteRequest,
  ExportPostmanCollection,
  FormatBody,
  GetState,
  ImportCurlRequest,
  ImportFetchRequest,
  ImportPostmanCollection,
  RunCollection,
  SaveCollection,
  SaveEnvironment,
  SaveGlobals,
  SaveRequest,
  SelectFile,
  SendRequest,
  SetActiveEnvironment,
  TestSSE,
  TestWebSocket
} from '../wailsjs/go/main/App'
import { domain, realtime } from '../wailsjs/go/models'
import AppTitleBar from './components/AppTitleBar.vue'
import EnvironmentsView from './components/EnvironmentsView.vue'
import HistoryView from './components/HistoryView.vue'
import ImportModal from './components/ImportModal.vue'
import RealtimeView from './components/RealtimeView.vue'
import RequestWorkspace from './components/RequestWorkspace.vue'
import RunnerView from './components/RunnerView.vue'
import SettingsView from './components/SettingsView.vue'
import SidebarRail from './components/SidebarRail.vue'
import WorkspaceSidebar from './components/WorkspaceSidebar.vue'
import { authTypes, bodyModes, methods } from './constants/request'
import { messages } from './i18n/messages'
import type { ActiveModal, JsonToken, Language, NavKey, RequestTab, ResponseTab, ResponseView, Theme } from './types'
import { formatError } from './utils/format'
import { tokenizeJSON } from './utils/jsonHighlight'

const state = ref<domain.WorkspaceState | null>(null)
const language = ref<Language>((localStorage.getItem('restdeck.language') as Language) || 'zh-CN')
const theme = ref<Theme>((localStorage.getItem('restdeck.theme') as Theme) || 'light')
const activeNav = ref<NavKey>('collections')
const activeRequestTab = ref<RequestTab>('params')
const activeResponseTab = ref<ResponseTab>('body')
const responseView = ref<ResponseView>('pretty')
const activeCollectionId = ref('')
const activeRequest = ref<domain.Request | null>(null)
const response = ref<domain.Response | null>(null)
const search = ref('')
const busy = ref(false)
const runnerBusy = ref(false)
const realtimeBusy = ref(false)
const statusMessage = ref('')
const activeModal = ref<ActiveModal>(null)
const postmanText = ref('')
const fetchText = ref('')
const curlText = ref('')
const collectionPickerOpen = ref(false)
const addMenuOpen = ref(false)
const optionsMenuOpen = ref(false)
const collectionToolbarRef = ref<HTMLElement | null>(null)
const editingCollectionId = ref('')
const editingCollectionName = ref('')
const pendingDeleteCollectionId = ref('')
const exportText = ref('')
const runnerResult = ref<domain.RunnerResult | null>(null)
const wsDraft = reactive({
  url: 'wss://echo.websocket.events',
  message: '{ "hello": "restdeck" }',
  headers: [] as domain.KeyValue[],
  timeoutMs: 10000
})
const sseDraft = reactive({
  url: '{{baseUrl}}/sse',
  headers: [] as domain.KeyValue[],
  timeoutMs: 10000,
  maxEvents: 5
})
const wsResult = ref<realtime.WebSocketResult | null>(null)
const sseResult = ref<realtime.SSEResult | null>(null)

const envDraft = reactive({
  id: '',
  name: '',
  variables: [] as domain.KeyValue[]
})
const globalsDraft = ref<domain.KeyValue[]>([])
const settingsDraft = reactive<{ language: Language; theme: Theme }>({ language: 'zh-CN', theme: 'light' })
const t = computed(() => messages[language.value])
const navItems = computed(() => [
  { key: 'collections' as NavKey, label: t.value.collections, icon: ListTree },
  { key: 'environments' as NavKey, label: t.value.environments, icon: Globe2 },
  { key: 'history' as NavKey, label: t.value.history, icon: History },
  { key: 'runner' as NavKey, label: t.value.runner, icon: Play },
  { key: 'realtime' as NavKey, label: t.value.realtime, icon: Radio },
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

const requestTabs = computed(() => [
  { key: 'params' as RequestTab, label: t.value.params, count: activeRequest.value?.params?.filter((item) => item.enabled && item.key).length ?? 0 },
  { key: 'auth' as RequestTab, label: t.value.auth, count: authBadgeCount(activeRequest.value) },
  { key: 'headers' as RequestTab, label: t.value.headers, count: activeRequest.value?.headers?.filter((item) => item.enabled && item.key).length ?? 0 },
  { key: 'body' as RequestTab, label: t.value.body, count: activeRequest.value?.bodyMode && activeRequest.value.bodyMode !== 'none' ? 1 : 0 },
  { key: 'pre' as RequestTab, label: t.value.pre, count: activeRequest.value?.preScript?.trim() ? 1 : 0 },
  { key: 'tests' as RequestTab, label: t.value.tests, count: activeRequest.value?.testScript?.trim() ? 1 : 0 },
  { key: 'settings' as RequestTab, label: t.value.settings, count: activeRequest.value?.timeoutMs ? 1 : 0 }
])

const responseTabs = computed(() => [
  { key: 'body' as ResponseTab, label: t.value.body, count: response.value?.body ? 1 : 0 },
  { key: 'headers' as ResponseTab, label: t.value.headers, count: response.value?.headers?.length ?? 0 },
  { key: 'cookies' as ResponseTab, label: t.value.cookies, count: response.value?.cookies?.length ?? 0 },
  { key: 'tests' as ResponseTab, label: t.value.tests, count: response.value?.testResults?.length ?? 0 }
])

onMounted(async () => {
  settingsDraft.language = language.value
  settingsDraft.theme = theme.value
  document.addEventListener('click', handleDocumentClick, true)
  await loadState()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick, true)
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
  envDraft.id = env.id
  envDraft.name = env.name
  envDraft.variables = cloneKeyValues(env.variables ?? [])
}, { immediate: true })

watch(state, (next) => {
  globalsDraft.value = cloneKeyValues(next?.globals ?? [])
})

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
}

function selectRequest(request: domain.Request) {
  activeRequest.value = normalizeRequest(cloneRequest(request))
  response.value = null
  activeResponseTab.value = 'body'
}

function selectHistoryRequest(request: domain.Request) {
  selectRequest(request)
  activeNav.value = 'collections'
}

async function createCollection() {
  const name = `${t.value.collections} ${(state.value?.collections?.length ?? 0) + 1}`
  const next = await CreateCollection(name)
  setState(next)
  activeCollectionId.value = next.collections[next.collections.length - 1]?.id ?? activeCollectionId.value
  statusMessage.value = t.value.collectionCreated
}

function selectCollection(collection: domain.Collection) {
  activeCollectionId.value = collection.id
  activeRequest.value = collection.requests?.[0] ? normalizeRequest(cloneRequest(collection.requests[0])) : null
  response.value = null
  collectionPickerOpen.value = false
  pendingDeleteCollectionId.value = ''
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

function createRequest() {
  const collection = activeCollection.value
  if (!collection) return
  activeRequest.value = new domain.Request({
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
    preScript: '',
    testScript: 'pm.test("Status is successful", function () { expect(pm.response.code).to.be.ok(); });',
    timeoutMs: 30000,
    sortOrder: collection.requests?.length ?? 0
  })
  response.value = null
  addMenuOpen.value = false
}

async function saveActiveRequest() {
  if (!activeRequest.value) return
  syncFormBody(activeRequest.value)
  busy.value = true
  try {
    const next = await SaveRequest(activeRequest.value)
    setState(next)
    statusMessage.value = t.value.requestSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function deleteActiveRequest() {
  if (!activeRequest.value?.id) return
  busy.value = true
  try {
    const next = await DeleteRequest(activeRequest.value.id)
    activeRequest.value = null
    response.value = null
    setState(next)
    statusMessage.value = t.value.requestDeleted
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

async function sendActiveRequest() {
  if (!activeRequest.value) return
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

async function importPostmanCollection() {
  if (!postmanText.value.trim()) return
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

async function importFetchRequest() {
  if (!fetchText.value.trim()) return
  busy.value = true
  try {
    const previousRequestIds = new Set(activeCollection.value?.requests?.map((request) => request.id) ?? [])
    const collectionID = activeCollection.value?.id ?? ''
    const next = await ImportFetchRequest(fetchText.value, collectionID)
    setState(next)
    selectLatestImportedRequest(previousRequestIds)
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
  busy.value = true
  try {
    const previousRequestIds = new Set(activeCollection.value?.requests?.map((request) => request.id) ?? [])
    const collectionID = activeCollection.value?.id ?? ''
    const next = await ImportCurlRequest(curlText.value, collectionID)
    setState(next)
    selectLatestImportedRequest(previousRequestIds)
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

function selectLatestImportedRequest(previousRequestIds: Set<string>) {
  const imported = activeCollection.value?.requests?.find((request) => !previousRequestIds.has(request.id))
  if (imported) selectRequest(imported)
}

function openPostmanModal() {
  closeCollectionMenus()
  postmanText.value = JSON.stringify({ info: { name: 'Imported' }, item: [] }, null, 2)
  activeModal.value = 'postman'
}

function openFetchModal() {
  closeCollectionMenus()
  fetchText.value = `fetch("https://api.example.com/v1/resource", {
  "headers": {
    "accept": "application/json"
  },
  "method": "GET"
});`
  activeModal.value = 'fetch'
}

function openCurlModal() {
  closeCollectionMenus()
  curlText.value = `curl 'https://api.example.com/v1/resource' \\
  -H 'accept: application/json'`
  activeModal.value = 'curl'
}

function closeModal() {
  activeModal.value = null
}

function submitActiveModal() {
  if (activeModal.value === 'postman') return importPostmanCollection()
  if (activeModal.value === 'fetch') return importFetchRequest()
  if (activeModal.value === 'curl') return importCurlRequest()
}

async function exportCollection() {
  const collection = activeCollection.value
  if (!collection) return
  try {
    exportText.value = await ExportPostmanCollection(collection.id)
    activeModal.value = 'export'
    closeCollectionMenus()
    statusMessage.value = t.value.collectionExported
  } catch (error) {
    statusMessage.value = formatError(error)
  }
}

function closeCollectionMenus() {
  collectionPickerOpen.value = false
  addMenuOpen.value = false
  optionsMenuOpen.value = false
}

function handleDocumentClick(event: MouseEvent) {
  if (!collectionPickerOpen.value && !addMenuOpen.value && !optionsMenuOpen.value) return
  const target = event.target
  if (target instanceof Node && collectionToolbarRef.value?.contains(target)) return
  closeCollectionMenus()
  pendingDeleteCollectionId.value = ''
}

async function saveEnvironmentDraft() {
  const env = new domain.Environment({
    id: envDraft.id,
    name: envDraft.name || t.value.environments,
    variables: envDraft.variables,
    isActive: true
  })
  const next = await SaveEnvironment(env)
  setState(next)
  statusMessage.value = t.value.environmentSaved
}

async function setEnvironment(id: string) {
  const next = await SetActiveEnvironment(id)
  setState(next)
  statusMessage.value = t.value.environmentSelected
}

async function saveGlobalsDraft() {
  const next = await SaveGlobals(globalsDraft.value)
  setState(next)
  statusMessage.value = t.value.globalsSaved
}

async function runActiveCollection() {
  const collection = activeCollection.value
  if (!collection) return
  runnerBusy.value = true
  runnerResult.value = null
  try {
    runnerResult.value = await RunCollection(collection.id, activeEnvironment.value?.id ?? '', 1)
    statusMessage.value = `${t.value.runner}: ${runnerResult.value.passed} passed, ${runnerResult.value.failed} failed`
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    runnerBusy.value = false
  }
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

function setAuthType(value: string) {
  if (!activeRequest.value) return
  activeRequest.value.auth = new domain.AuthConfig({ type: value, values: defaultAuthValues(value) })
}

function defaultAuthValues(type: string) {
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

function authBadgeCount(request: domain.Request | null) {
  if (!request?.auth?.type || request.auth.type === 'none') return 0
  return 1
}

function newKeyValue() {
  return new domain.KeyValue({ id: crypto.randomUUID(), enabled: true, key: '', value: '', description: '', secret: false })
}

function newFormItem() {
  return new domain.FormItem({ id: crypto.randomUUID(), enabled: true, key: '', type: 'text', value: '', filePath: '', description: '' })
}

function cloneKeyValues(items: domain.KeyValue[]) {
  return items.map((item) => new domain.KeyValue({ ...item }))
}

function cloneRequest(request: domain.Request) {
  return new domain.Request(JSON.parse(JSON.stringify(request)))
}

function normalizeRequest(request: domain.Request) {
  request.params = request.params ?? []
  request.headers = request.headers ?? []
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

function normalizeFormItems(items: domain.FormItem[], fallbackBody: string) {
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

function formItemsFromBody(raw: string) {
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

function formItemsToBody(items: domain.FormItem[]) {
  return (items ?? [])
    .filter((item) => item.key || item.value || item.filePath)
    .map((item) => `${item.key}=${item.type === 'file' ? `@${item.filePath}` : item.value}`)
    .join('\n')
}

function syncFormBody(request: domain.Request) {
  if (request.bodyMode === 'form') {
    request.formItems = normalizeFormItems(request.formItems ?? [], request.body ?? '')
    request.body = formItemsToBody(request.formItems)
  } else {
    request.formItems = normalizeFormItems(request.formItems ?? [], '')
  }
}

function minimiseWindow() {
  WindowMinimise()
}

function toggleWindowMaximise() {
  WindowToggleMaximise()
}

function closeWindow() {
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

    <main class="workspace">
      <SidebarRail v-model:active-nav="activeNav" :items="navItems" />

      <WorkspaceSidebar
        v-model:collection-picker-open="collectionPickerOpen"
        v-model:add-menu-open="addMenuOpen"
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
        :history="state?.history ?? []"
        :editing-collection-id="editingCollectionId"
        :pending-delete-collection-id="pendingDeleteCollectionId"
        :runner-busy="runnerBusy"
        @set-toolbar-ref="collectionToolbarRef = $event"
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
        @export-collection="exportCollection"
        @select-request="selectRequest"
        @select-environment="setEnvironment"
        @run-collection="runActiveCollection"
      />

      <section class="main-pane">
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
          :busy="busy"
          :methods="methods"
          :auth-types="authTypes"
          :body-modes="bodyModes"
          :send-request-action="sendActiveRequest"
          @save-request="saveActiveRequest"
          @delete-request="deleteActiveRequest"
          @export-collection="exportCollection"
          @create-request="createRequest"
          @add-param="addParam"
          @add-header="addHeader"
          @add-form-item="addFormItem"
          @remove-row="removeRow"
          @remove-form-item="removeFormItem"
          @select-form-file="selectFormFile"
          @set-auth-type="setAuthType"
        />

        <EnvironmentsView
          v-else-if="activeNav === 'environments'"
          v-model:env-draft="envDraft"
          v-model:globals-draft="globalsDraft"
          :t="t"
          :active-environment="activeEnvironment"
          @save-environment="saveEnvironmentDraft"
          @save-globals="saveGlobalsDraft"
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
          :active-collection="activeCollection"
          :runner-result="runnerResult"
          :runner-busy="runnerBusy"
          @run-collection="runActiveCollection"
        />

        <RealtimeView
          v-else-if="activeNav === 'realtime'"
          v-model:ws-draft="wsDraft"
          v-model:sse-draft="sseDraft"
          :t="t"
          :realtime-busy="realtimeBusy"
          :ws-result="wsResult"
          :sse-result="sseResult"
          @run-web-socket="runWebSocketCheck"
          @run-s-s-e="runSSECheck"
        />

        <SettingsView
          v-else
          v-model:language="language"
          :t="t"
        />
      </section>
    </main>

    <footer class="statusbar">
      <span>{{ statusMessage }}</span>
    </footer>
  </div>

  <ImportModal
    v-model:postman-text="postmanText"
    v-model:fetch-text="fetchText"
    v-model:curl-text="curlText"
    :active-modal="activeModal"
    :title="activeModalTitle"
    :busy="busy"
    :t="t"
    :export-text="exportText"
    @close="closeModal"
    @submit="submitActiveModal"
  />
</template>
