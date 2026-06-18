<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import {
  Activity,
  CheckCircle2,
  Clock3,
  CircleAlert,
  ChevronDown,
  Download,
  FileJson2,
  Globe2,
  Home,
  History,
  Import,
  KeyRound,
  ListTree,
  Loader2,
  MoreHorizontal,
  Pencil,
  Play,
  Plus,
  Save,
  Search,
  Send,
  Settings,
  Square,
  Trash2,
  Radio,
  X,
  XCircle
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
  SendRequest,
  SetActiveEnvironment,
  TestSSE,
  TestWebSocket
} from '../wailsjs/go/main/App'
import { domain, realtime } from '../wailsjs/go/models'

type NavKey = 'collections' | 'environments' | 'history' | 'runner' | 'realtime' | 'settings'
type RequestTab = 'params' | 'auth' | 'headers' | 'body' | 'pre' | 'tests' | 'settings'
type ResponseTab = 'body' | 'headers' | 'cookies' | 'tests'
type ResponseView = 'pretty' | 'raw' | 'preview'
type ActiveModal = 'postman' | 'fetch' | 'curl' | 'export' | null
type Language = 'zh-CN' | 'en-US'
type JsonTokenType = 'plain' | 'key' | 'string' | 'number' | 'boolean' | 'null' | 'punctuation'
type JsonToken = { type: JsonTokenType; text: string }

const messages = {
  'zh-CN': {
    home: '主页',
    workspace: 'RestDeck',
    workspaceLoaded: '工作区已加载',
    new: '新建',
    import: '导入',
    export: '导出',
    save: '保存',
    send: '发送',
    run: '运行',
    listen: '监听',
    connect: '连接',
    search: '搜索',
    collections: '集合',
    environments: '环境',
    history: '历史',
    runner: 'Runner',
    realtime: '实时',
    settings: '设置',
    params: '参数',
    auth: '认证',
    headers: '请求头',
    body: 'Body',
    pre: '预请求',
    tests: '测试',
    response: '响应',
    cookies: 'Cookies',
    status: '状态',
    duration: '耗时',
    size: '大小',
    key: '键',
    value: '值',
    description: '描述',
    addParam: '添加参数',
    addHeader: '添加请求头',
    addVariable: '添加变量',
    addGlobal: '添加全局变量',
    newRequest: '新建请求',
    newCollection: '新建集合',
    editCollection: '编辑集合',
    deleteCollection: '删除集合',
    confirmDeleteCollection: '再次点击删除',
    collectionDeleted: '集合已删除',
    type: '类型',
    addTo: '添加到',
    username: '用户名',
    password: '密码',
    consumerKey: 'Consumer Key',
    consumerSecret: 'Consumer Secret',
    tokenSecret: 'Token Secret',
    accessToken: 'Access Token',
    timeout: '超时',
    passed: '通过',
    failed: '失败',
    maxEvents: '最大事件数',
    collectionName: '集合名称',
    importCollection: '导入请求',
    importPostmanCollection: '导入 Postman 集合',
    importFetchRequest: '导入 Fetch 请求',
    importCurlRequest: '导入 cURL 请求',
    importFromFetch: '从 Fetch 导入',
    importFromCurl: '从 cURL 导入',
    importFromPostman: '从 Postman 导入',
    collectionOptions: '集合选项',
    postmanJSON: 'Postman JSON',
    fetchSnippet: 'Fetch',
    curlSnippet: 'cURL',
    collectionSaved: '集合已保存',
    exportedCollection: '导出的集合',
    requestName: '未命名请求',
    requestSaved: '请求已保存',
    requestDeleted: '请求已删除',
    collectionCreated: '集合已创建',
    collectionImported: 'Postman 集合已导入',
    fetchImported: 'Fetch 请求已导入',
    curlImported: 'cURL 请求已导入',
    collectionExported: '集合已导出',
    environmentSaved: '环境已保存',
    environmentSelected: '环境已切换',
    globalsSaved: '全局变量已保存',
    noBody: '此请求不发送 Body。',
    noResponse: '发送请求后在这里查看响应。',
    noCookies: '响应没有返回 Cookie。',
    noTests: '没有测试脚本结果。',
    createOrSelect: '新建或选择一个请求开始。',
    runnerHelp: '使用当前环境运行所选集合一次。',
    runnerTitle: '集合 Runner',
    runnerEmpty: 'Runner 结果会显示在这里。',
    realtimeHelp: '使用当前环境变量调试 WebSocket 和 Server-Sent Events。',
    realtimeTitle: '实时客户端',
    realtimeSubtitle: '这里支持 WebSocket 和 SSE；gRPC 未实现前不会显示入口。',
    websocketDesc: '连接、发送一条消息并读取响应。',
    sseDesc: '连接事件流并采集事件。',
    wsEmpty: 'WebSocket 结果会显示在这里。',
    sseEmpty: 'SSE 事件会显示在这里。',
    connected: '已连接',
    received: '收到',
    event: '事件',
    error: '错误',
    settingsHelp: '仅显示本地应用设置。',
    settingsSide: '仅本地偏好设置。账号、云同步和团队设置没有实现，因此不显示。',
    language: '语言',
    theme: '主题',
    light: '浅色',
    localOnly: '仅本地',
    honestUI: '真实 UI',
    encryptedNote: '请求认证和环境密钥会在本地加密保存。',
    noAuth: '此请求不会添加认证头。',
    activeEnvironment: '当前环境',
    pretty: '美化',
    raw: '原始',
    preview: '预览'
  },
  'en-US': {
    home: 'Home',
    workspace: 'RestDeck',
    workspaceLoaded: 'Workspace loaded',
    new: 'New',
    import: 'Import',
    export: 'Export',
    save: 'Save',
    send: 'Send',
    run: 'Run',
    listen: 'Listen',
    connect: 'Connect',
    search: 'Search',
    collections: 'Collections',
    environments: 'Environments',
    history: 'History',
    runner: 'Runner',
    realtime: 'Realtime',
    settings: 'Settings',
    params: 'Params',
    auth: 'Auth',
    headers: 'Headers',
    body: 'Body',
    pre: 'Pre-request',
    tests: 'Tests',
    response: 'Response',
    cookies: 'Cookies',
    status: 'Status',
    duration: 'Duration',
    size: 'Size',
    key: 'KEY',
    value: 'VALUE',
    description: 'DESCRIPTION',
    addParam: 'Add param',
    addHeader: 'Add header',
    addVariable: 'Add variable',
    addGlobal: 'Add global',
    newRequest: 'New request',
    newCollection: 'New collection',
    editCollection: 'Edit collection',
    deleteCollection: 'Delete collection',
    confirmDeleteCollection: 'Click again to delete',
    collectionDeleted: 'Collection deleted',
    type: 'Type',
    addTo: 'Add to',
    username: 'Username',
    password: 'Password',
    consumerKey: 'Consumer key',
    consumerSecret: 'Consumer secret',
    tokenSecret: 'Token secret',
    accessToken: 'Access token',
    timeout: 'Timeout',
    passed: 'Passed',
    failed: 'Failed',
    maxEvents: 'Max events',
    collectionName: 'Collection name',
    importCollection: 'Import request',
    importPostmanCollection: 'Import Postman Collection',
    importFetchRequest: 'Import Fetch Request',
    importCurlRequest: 'Import cURL Request',
    importFromFetch: 'Import from Fetch',
    importFromCurl: 'Import from cURL',
    importFromPostman: 'Import from Postman',
    collectionOptions: 'Collection options',
    postmanJSON: 'Postman JSON',
    fetchSnippet: 'Fetch',
    curlSnippet: 'cURL',
    collectionSaved: 'Collection saved',
    exportedCollection: 'Exported Collection',
    requestName: 'Untitled Request',
    requestSaved: 'Request saved',
    requestDeleted: 'Request deleted',
    collectionCreated: 'Collection created',
    collectionImported: 'Postman collection imported',
    fetchImported: 'Fetch request imported',
    curlImported: 'cURL request imported',
    collectionExported: 'Collection exported',
    environmentSaved: 'Environment saved',
    environmentSelected: 'Environment selected',
    globalsSaved: 'Globals saved',
    noBody: 'No request body for this method.',
    noResponse: 'Send a request to inspect the response.',
    noCookies: 'No cookies returned.',
    noTests: 'No test script results.',
    createOrSelect: 'Create or select a request to start.',
    runnerHelp: 'Run the selected collection once with the active environment.',
    runnerTitle: 'Collection runner',
    runnerEmpty: 'Runner results will appear here.',
    realtimeHelp: 'Debug WebSocket and Server-Sent Events with the active environment variables.',
    realtimeTitle: 'Realtime clients',
    realtimeSubtitle: 'WebSocket and SSE are available here; gRPC is intentionally absent until implemented.',
    websocketDesc: 'Connect, send one message, read replies.',
    sseDesc: 'Connect and collect event-stream messages.',
    wsEmpty: 'WebSocket result will appear here.',
    sseEmpty: 'SSE events will appear here.',
    connected: 'Connected',
    received: 'Received',
    event: 'Event',
    error: 'Error',
    settingsHelp: 'Only local application settings are shown.',
    settingsSide: 'Local preferences only. Account, cloud sync and team settings are intentionally absent.',
    language: 'Language',
    theme: 'Theme',
    light: 'Light',
    localOnly: 'Local only',
    honestUI: 'Honest UI',
    encryptedNote: 'Sensitive request auth and secret environment values are encrypted locally.',
    noAuth: 'This request does not add authorization headers.',
    activeEnvironment: 'Active environment',
    pretty: 'Pretty',
    raw: 'Raw',
    preview: 'Preview'
  }
}

const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']
const authTypes = [
  { value: 'none', label: 'No Auth' },
  { value: 'apiKey', label: 'API Key' },
  { value: 'bearer', label: 'Bearer Token' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'digest', label: 'Digest Auth' },
  { value: 'oauth1', label: 'OAuth 1.0' },
  { value: 'oauth2', label: 'OAuth 2.0 Token' }
]
const bodyModes = [
  { value: 'none', label: 'None' },
  { value: 'json', label: 'JSON' },
  { value: 'raw', label: 'Raw' },
  { value: 'urlencoded', label: 'x-www-form-urlencoded' },
  { value: 'form', label: 'Form data' }
]

const state = ref<domain.WorkspaceState | null>(null)
const language = ref<Language>((localStorage.getItem('restdeck.language') as Language) || 'zh-CN')
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
const collectionDraft = reactive({ name: '' })
const settingsDraft = reactive({ language: 'zh-CN', theme: 'light' })
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

watch(activeEnvironment, (env) => {
  if (!env) return
  envDraft.id = env.id
  envDraft.name = env.name
  envDraft.variables = cloneKeyValues(env.variables ?? [])
}, { immediate: true })

watch(activeCollection, (collection) => {
  collectionDraft.name = collection?.name ?? ''
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
    if (first) activeRequest.value = cloneRequest(first)
  } else {
    const fresh = collection?.requests?.find((request) => request.id === activeRequest.value?.id)
    if (fresh) {
      activeRequest.value = cloneRequest(fresh)
    } else {
      activeRequest.value = collection?.requests?.[0] ? cloneRequest(collection.requests[0]) : null
      response.value = null
    }
  }
}

function selectRequest(request: domain.Request) {
  activeRequest.value = cloneRequest(request)
  response.value = null
  activeResponseTab.value = 'body'
}

async function createCollection() {
  const name = `${t.value.collections} ${(state.value?.collections?.length ?? 0) + 1}`
  const next = await CreateCollection(name)
  setState(next)
  activeCollectionId.value = next.collections[next.collections.length - 1]?.id ?? activeCollectionId.value
  statusMessage.value = t.value.collectionCreated
}

async function saveActiveCollection() {
  const collection = activeCollection.value
  if (!collection) return
  const id = collection.id
  busy.value = true
  try {
    const next = await SaveCollection(new domain.Collection({
      ...collection,
      name: collectionDraft.name.trim() || collection.name
    }))
    setState(next)
    activeCollectionId.value = id
    statusMessage.value = t.value.collectionSaved
  } catch (error) {
    statusMessage.value = formatError(error)
  } finally {
    busy.value = false
  }
}

function selectCollection(collection: domain.Collection) {
  activeCollectionId.value = collection.id
  activeRequest.value = collection.requests?.[0] ? cloneRequest(collection.requests[0]) : null
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
      activeRequest.value = next.collections?.[0]?.requests?.[0] ? cloneRequest(next.collections[0].requests[0]) : null
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
    id: '',
    collectionId: collection.id,
    parentId: '',
    name: t.value.requestName,
    method: 'GET',
    url: '{{baseUrl}}/anything',
    params: [newKeyValue()],
    headers: [new domain.KeyValue({ id: crypto.randomUUID(), enabled: true, key: 'Accept', value: 'application/json', description: '', secret: false })],
    bodyMode: 'none',
    body: '',
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
  busy.value = true
  response.value = null
  try {
    const savedState = await SaveRequest(activeRequest.value)
    setState(savedState)
    const result = await SendRequest(activeRequest.value, activeEnvironment.value?.id ?? '', globalsDraft.value)
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

function addVariable(target: domain.KeyValue[]) {
  target.push(newKeyValue())
}

function removeRow(target: domain.KeyValue[], index: number) {
  target.splice(index, 1)
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

function cloneKeyValues(items: domain.KeyValue[]) {
  return items.map((item) => new domain.KeyValue({ ...item }))
}

function cloneRequest(request: domain.Request) {
  return new domain.Request(JSON.parse(JSON.stringify(request)))
}

function formatError(error: unknown) {
  if (error instanceof Error) return error.message
  return String(error)
}

function formatBytes(bytes?: number) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function statusClass(code?: number) {
  if (!code) return 'text-zinc-500'
  if (code >= 200 && code < 300) return 'text-emerald-600'
  if (code >= 300 && code < 400) return 'text-sky-600'
  if (code >= 400) return 'text-red-600'
  return 'text-zinc-600'
}

function responseStatusText(item: domain.Response) {
  return item.status || String(item.statusCode || '')
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

function tokenizeJSON(raw: string): JsonToken[] {
  const tokens: JsonToken[] = []
  const pattern = /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(?=\s*:))|("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|\b(true|false)\b|\bnull\b|([{}[\],:])/g
  let lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = pattern.exec(raw)) !== null) {
    if (match.index > lastIndex) {
      tokens.push({ type: 'plain', text: raw.slice(lastIndex, match.index) })
    }
    const text = match[0]
    let type: JsonTokenType = 'plain'
    if (match[1]) type = 'key'
    else if (match[2]) type = 'string'
    else if (match[3]) type = 'number'
    else if (match[4]) type = 'boolean'
    else if (text === 'null') type = 'null'
    else if (match[5]) type = 'punctuation'
    tokens.push({ type, text })
    lastIndex = pattern.lastIndex
  }
  if (lastIndex < raw.length) {
    tokens.push({ type: 'plain', text: raw.slice(lastIndex) })
  }
  return tokens
}

</script>

<template>
  <div class="app-shell">
    <header class="topbar window-titlebar" @dblclick="toggleWindowMaximise">
      <div class="window-title">RestDeck</div>
      <button class="top-link" @dblclick.stop @click="activeNav = 'collections'">
        <Home :size="14" />
        {{ t.home }}
      </button>
      <div class="top-search" @dblclick.stop>
        <Search :size="14" />
        <input v-model="search" :placeholder="t.search" />
      </div>
      <div class="top-spacer" />
      <select class="env-select" :value="activeEnvironment?.id" @dblclick.stop @change="setEnvironment(($event.target as HTMLSelectElement).value)">
        <option v-for="env in state?.environments ?? []" :key="env.id" :value="env.id">{{ env.name }}</option>
      </select>
      <div class="window-controls" @dblclick.stop>
        <button type="button" class="window-control" title="Minimize" @click="minimiseWindow">
          <span class="minimize-mark"></span>
        </button>
        <button type="button" class="window-control" title="Maximize" @click="toggleWindowMaximise">
          <Square :size="11" :stroke-width="1.7" />
        </button>
        <button type="button" class="window-control close" title="Close" @click="closeWindow">
          <X :size="14" :stroke-width="1.7" />
        </button>
      </div>
    </header>

    <main class="workspace">
      <aside class="rail">
        <button
          v-for="item in navItems"
          :key="item.key"
          :class="['rail-button', { active: activeNav === item.key }]"
          :title="item.label"
          @click="activeNav = item.key"
        >
          <component :is="item.icon" :size="17" />
          <span>{{ item.label }}</span>
        </button>
      </aside>

      <aside class="sidebar">
        <template v-if="activeNav === 'collections'">
          <div ref="collectionToolbarRef" class="collection-toolbar">
            <div class="collection-picker-wrap">
              <button class="collection-link" type="button" @click="collectionPickerOpen = !collectionPickerOpen; addMenuOpen = false; optionsMenuOpen = false">
                <span class="truncate">{{ activeCollection?.name ?? t.collections }}</span>
                <ChevronDown :size="13" />
              </button>
              <div v-if="collectionPickerOpen" class="collection-dropdown">
                <div class="collection-dropdown-list">
                  <div
                    v-for="collection in state?.collections ?? []"
                    :key="collection.id"
                    :class="['collection-option', { active: collection.id === activeCollection?.id }]"
                  >
                    <input
                      v-if="editingCollectionId === collection.id"
                      v-model="editingCollectionName"
                      class="collection-rename-input"
                      :placeholder="t.collectionName"
                      @keydown.enter="saveEditingCollection(collection)"
                      @keydown.esc="cancelEditingCollection"
                      @blur="saveEditingCollection(collection)"
                    />
                    <button v-else class="collection-option-name" type="button" @click="selectCollection(collection)">
                      <span class="truncate">{{ collection.name }}</span>
                    </button>
                    <button class="ghost-icon" type="button" :title="t.editCollection" @click.stop="startEditingCollection(collection)">
                      <Pencil :size="13" />
                    </button>
                    <button
                      :class="['ghost-icon', 'danger-icon', { pending: pendingDeleteCollectionId === collection.id }]"
                      type="button"
                      :title="pendingDeleteCollectionId === collection.id ? t.confirmDeleteCollection : t.deleteCollection"
                      @click.stop="deleteCollectionFromPicker(collection)"
                    >
                      <CircleAlert v-if="pendingDeleteCollectionId === collection.id" :size="13" />
                      <X v-else :size="13" />
                    </button>
                  </div>
                </div>
                <button class="collection-new-option" type="button" @click="createCollection">
                  <Plus :size="14" />
                  {{ t.newCollection }}
                </button>
              </div>
            </div>

            <div class="collection-actions">
              <div class="menu-wrap">
                <button class="icon-btn" type="button" :title="t.new" @click="addMenuOpen = !addMenuOpen; collectionPickerOpen = false; optionsMenuOpen = false">
                  <Plus :size="15" />
                </button>
                <div v-if="addMenuOpen" class="action-menu">
                  <button type="button" @click="createRequest">
                    <Plus :size="14" />
                    {{ t.newRequest }}
                  </button>
                  <button type="button" @click="openFetchModal">
                    <Import :size="14" />
                    {{ t.importFromFetch }}
                  </button>
                  <button type="button" @click="openCurlModal">
                    <Import :size="14" />
                    {{ t.importFromCurl }}
                  </button>
                </div>
              </div>
              <div class="menu-wrap">
                <button class="icon-btn" type="button" :title="t.collectionOptions" @click="optionsMenuOpen = !optionsMenuOpen; collectionPickerOpen = false; addMenuOpen = false">
                  <MoreHorizontal :size="14" />
                </button>
                <div v-if="optionsMenuOpen" class="action-menu right">
                  <button type="button" @click="openPostmanModal">
                    <Import :size="14" />
                    {{ t.importFromPostman }}
                  </button>
                  <button type="button" :disabled="!activeCollection" @click="exportCollection">
                    <Download :size="14" />
                    {{ t.export }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </template>
        <div v-else class="sidebar-title">
          <span>{{ navItems.find((item) => item.key === activeNav)?.label }}</span>
        </div>

        <template v-if="activeNav === 'collections'">
          <div class="request-list">
            <button
              v-for="request in filteredRequests"
              :key="request.id"
              :class="['request-row', { active: request.id === activeRequest?.id }]"
              @click="selectRequest(request)"
            >
              <span :class="['method', request.method.toLowerCase()]">{{ request.method }}</span>
              <span class="truncate">{{ request.name }}</span>
            </button>
            <div v-if="!activeCollection" class="side-note">{{ t.createOrSelect }}</div>
          </div>
        </template>

        <template v-else-if="activeNav === 'environments'">
          <div class="request-list">
            <button
              v-for="env in state?.environments ?? []"
              :key="env.id"
              :class="['request-row', { active: env.id === activeEnvironment?.id }]"
              @click="setEnvironment(env.id)"
            >
              <Globe2 :size="14" />
              <span class="truncate">{{ env.name }}</span>
            </button>
          </div>
        </template>

        <template v-else-if="activeNav === 'history'">
          <div class="request-list">
            <button v-for="item in state?.history ?? []" :key="item.id" class="request-row" @click="selectRequest(item.request)">
              <span :class="['method', item.method.toLowerCase()]">{{ item.method }}</span>
              <span class="truncate">{{ item.name || item.url }}</span>
              <span :class="['history-code', statusClass(item.statusCode)]">{{ item.statusCode || '-' }}</span>
            </button>
          </div>
        </template>

        <template v-else-if="activeNav === 'runner'">
          <div class="side-note">{{ t.runnerHelp }}</div>
          <button class="primary-wide" :disabled="runnerBusy || !activeCollection" @click="runActiveCollection">
            <Loader2 v-if="runnerBusy" class="spin" :size="14" />
            <Play v-else :size="14" />
            Run collection
          </button>
        </template>

        <template v-else-if="activeNav === 'realtime'">
          <div class="side-note">{{ t.realtimeHelp }}</div>
          <button class="request-row active" type="button">
            <Radio :size="14" />
            <span>WebSocket</span>
          </button>
          <button class="request-row" type="button">
            <Activity :size="14" />
            <span>SSE</span>
          </button>
        </template>

        <template v-else>
          <div class="side-note">{{ t.settingsSide }}</div>
        </template>
      </aside>

      <section class="main-pane">
        <template v-if="activeNav === 'collections'">
          <div class="editor-header">
            <div class="breadcrumb">
              <span>{{ activeCollection?.name ?? 'Collection' }}</span>
              <span>/</span>
              <input v-if="activeRequest" v-model="activeRequest.name" class="title-input" />
            </div>
            <div class="editor-actions">
              <button class="toolbar-btn" :disabled="!activeRequest || busy" @click="saveActiveRequest">
                <Save :size="14" />
                {{ t.save }}
              </button>
              <button class="icon-btn" :disabled="!activeRequest?.id || busy" :title="t.requestDeleted" @click="deleteActiveRequest">
                <Trash2 :size="14" />
              </button>
              <button class="toolbar-btn" :disabled="!activeCollection" @click="exportCollection">
                <Download :size="14" />
                {{ t.export }}
              </button>
            </div>
          </div>

          <div v-if="activeRequest" class="request-line">
            <select v-model="activeRequest.method" class="method-select">
              <option v-for="method in methods" :key="method" :value="method">{{ method }}</option>
            </select>
            <input v-model="activeRequest.url" class="url-input" placeholder="https://api.example.com/v1/resource" />
            <button class="send-btn" :disabled="busy" @click="sendActiveRequest">
              <Loader2 v-if="busy" class="spin" :size="15" />
              <Send v-else :size="15" />
              {{ t.send }}
            </button>
          </div>

          <div v-if="activeRequest" class="split-editor">
            <section class="request-editor">
              <div class="tabs">
                <button
                  v-for="tab in requestTabs"
                  :key="tab.key"
                  :class="['tab', { active: activeRequestTab === tab.key }]"
                  @click="activeRequestTab = tab.key"
                >
                  {{ tab.label }}
                  <span v-if="tab.count" class="count">{{ tab.count }}</span>
                </button>
              </div>

              <div class="tab-panel">
                <div v-if="activeRequestTab === 'params'" class="kv-table">
                  <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
                  <div v-for="(param, index) in activeRequest.params" :key="param.id" class="kv-row">
                    <input v-model="param.enabled" type="checkbox" />
                    <input v-model="param.key" :placeholder="t.key" />
                    <input v-model="param.value" :placeholder="t.value" />
                    <input v-model="param.description" :placeholder="t.description" />
                    <button class="ghost-icon" @click="removeRow(activeRequest.params, index)"><Trash2 :size="13" /></button>
                  </div>
                  <button class="add-row" @click="addParam"><Plus :size="13" /> {{ t.addParam }}</button>
                </div>

                <div v-else-if="activeRequestTab === 'headers'" class="kv-table">
                  <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
                  <div v-for="(header, index) in activeRequest.headers" :key="header.id" class="kv-row">
                    <input v-model="header.enabled" type="checkbox" />
                    <input v-model="header.key" :placeholder="t.headers" />
                    <input v-model="header.value" :placeholder="t.value" />
                    <input v-model="header.description" :placeholder="t.description" />
                    <button class="ghost-icon" @click="removeRow(activeRequest.headers, index)"><Trash2 :size="13" /></button>
                  </div>
                  <button class="add-row" @click="addHeader"><Plus :size="13" /> {{ t.addHeader }}</button>
                </div>

                <div v-else-if="activeRequestTab === 'auth'" class="auth-grid">
                  <label>
                    <span>{{ t.type }}</span>
                    <select :value="activeRequest.auth?.type ?? 'none'" @change="setAuthType(($event.target as HTMLSelectElement).value)">
                      <option v-for="item in authTypes" :key="item.value" :value="item.value">{{ item.label }}</option>
                    </select>
                  </label>
                  <template v-if="activeRequest.auth?.type === 'apiKey'">
                    <label><span>{{ t.key }}</span><input v-model="activeRequest.auth.values.key" /></label>
                    <label><span>{{ t.value }}</span><input v-model="activeRequest.auth.values.value" type="password" /></label>
                    <label><span>{{ t.addTo }}</span><select v-model="activeRequest.auth.values.in"><option value="header">{{ t.headers }}</option><option value="query">{{ t.params }}</option></select></label>
                  </template>
                  <template v-else-if="activeRequest.auth?.type === 'bearer'">
                    <label><span>Token</span><input v-model="activeRequest.auth.values.token" type="password" /></label>
                  </template>
                  <template v-else-if="activeRequest.auth?.type === 'basic' || activeRequest.auth?.type === 'digest'">
                    <label><span>{{ t.username }}</span><input v-model="activeRequest.auth.values.username" /></label>
                    <label><span>{{ t.password }}</span><input v-model="activeRequest.auth.values.password" type="password" /></label>
                  </template>
                  <template v-else-if="activeRequest.auth?.type === 'oauth1'">
                    <label><span>{{ t.consumerKey }}</span><input v-model="activeRequest.auth.values.consumerKey" /></label>
                    <label><span>{{ t.consumerSecret }}</span><input v-model="activeRequest.auth.values.consumerSecret" type="password" /></label>
                    <label><span>Token</span><input v-model="activeRequest.auth.values.token" /></label>
                    <label><span>{{ t.tokenSecret }}</span><input v-model="activeRequest.auth.values.tokenSecret" type="password" /></label>
                  </template>
                  <template v-else-if="activeRequest.auth?.type === 'oauth2'">
                    <label><span>{{ t.accessToken }}</span><input v-model="activeRequest.auth.values.accessToken" type="password" /></label>
                  </template>
                  <p v-else class="muted">{{ t.noAuth }}</p>
                </div>

                <div v-else-if="activeRequestTab === 'body'" class="body-editor">
                  <select v-model="activeRequest.bodyMode" class="field compact">
                    <option v-for="mode in bodyModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
                  </select>
                  <textarea v-if="activeRequest.bodyMode !== 'none'" v-model="activeRequest.body" spellcheck="false" placeholder='{"hello": "world"}'></textarea>
                  <div v-else class="empty-panel">{{ t.noBody }}</div>
                </div>

                <div v-else-if="activeRequestTab === 'pre'" class="body-editor">
                  <textarea v-model="activeRequest.preScript" spellcheck="false" placeholder="pm.variables.set('traceId', '{{$guid}}');"></textarea>
                </div>

                <div v-else-if="activeRequestTab === 'tests'" class="body-editor">
                  <textarea v-model="activeRequest.testScript" spellcheck="false" placeholder='pm.test("Status is 200", function () { expect(pm.response.code).to.equal(200); });'></textarea>
                </div>

                <div v-else class="auth-grid">
                  <label><span>{{ t.timeout }} (ms)</span><input v-model.number="activeRequest.timeoutMs" type="number" min="1000" step="1000" /></label>
                </div>
              </div>
            </section>

            <section class="response-editor">
              <div class="response-meta">
                <strong>{{ t.response }}</strong>
                <span v-if="response" :class="statusClass(response.statusCode)">{{ responseStatusText(response) }}</span>
                <span v-if="response"><Clock3 :size="13" /> {{ response.durationMs }} ms</span>
                <span v-if="response">{{ formatBytes(response.sizeBytes) }}</span>
              </div>
              <div class="tabs">
                <button v-for="tab in responseTabs" :key="tab.key" :class="['tab', { active: activeResponseTab === tab.key }]" @click="activeResponseTab = tab.key">
                  {{ tab.label }}
                  <span v-if="tab.count" class="count">{{ tab.count }}</span>
                </button>
              </div>
              <div class="response-panel">
                <template v-if="!response">
                  <div class="empty-panel">{{ t.noResponse }}</div>
                </template>
                <template v-else-if="activeResponseTab === 'body'">
                  <div class="view-switch">
                    <button :class="{ active: responseView === 'pretty' }" @click="responseView = 'pretty'">{{ t.pretty }}</button>
                    <button :class="{ active: responseView === 'raw' }" @click="responseView = 'raw'">{{ t.raw }}</button>
                    <button :class="{ active: responseView === 'preview' }" @click="responseView = 'preview'">{{ t.preview }}</button>
                    <span class="pill">JSON</span>
                  </div>
                  <iframe v-if="responseView === 'preview'" :srcdoc="response.body" />
                  <pre v-else-if="responseView === 'pretty'" class="json-highlight"><span v-for="(token, index) in highlightedResponseBody" :key="index" :class="`json-${token.type}`">{{ token.text }}</span></pre>
                  <pre v-else>{{ prettyResponseBody }}</pre>
                </template>
                <template v-else-if="activeResponseTab === 'headers'">
                  <div class="kv-read" v-for="header in response.headers" :key="header.key"><span>{{ header.key }}</span><code>{{ header.value }}</code></div>
                </template>
                <template v-else-if="activeResponseTab === 'cookies'">
                  <div class="kv-read" v-for="cookie in response.cookies" :key="cookie.name"><span>{{ cookie.name }}</span><code>{{ cookie.value }}</code></div>
                  <div v-if="!response.cookies?.length" class="empty-panel">{{ t.noCookies }}</div>
                </template>
                <template v-else>
                  <div v-for="test in response.testResults" :key="test.name" class="test-row">
                    <CheckCircle2 v-if="test.passed" :size="15" class="text-emerald-600" />
                    <XCircle v-else :size="15" class="text-red-600" />
                    <span>{{ test.name }}</span>
                    <code v-if="test.message">{{ test.message }}</code>
                  </div>
                  <div v-if="!response.testResults?.length" class="empty-panel">{{ t.noTests }}</div>
                </template>
              </div>
            </section>
          </div>

          <div v-else class="blank-state">
            <FileJson2 :size="28" />
            <span>{{ t.createOrSelect }}</span>
            <button class="send-btn" @click="createRequest"><Plus :size="15" /> New request</button>
          </div>
        </template>

        <template v-else-if="activeNav === 'environments'">
          <div class="section-header">
            <div><h2>{{ t.environments }}</h2><p>{{ activeEnvironment?.name ?? t.environmentSelected }}</p></div>
            <button class="toolbar-btn" @click="saveEnvironmentDraft"><Save :size="14" /> {{ t.save }}</button>
          </div>
          <label class="stack-label"><span>{{ t.key }}</span><input v-model="envDraft.name" class="field" /></label>
          <div class="kv-table spacious">
            <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
            <div v-for="(variable, index) in envDraft.variables" :key="variable.id" class="kv-row">
              <input v-model="variable.enabled" type="checkbox" />
              <input v-model="variable.key" />
              <input v-model="variable.value" :type="variable.secret ? 'password' : 'text'" />
              <input v-model="variable.description" />
              <button class="ghost-icon" @click="removeRow(envDraft.variables, index)"><Trash2 :size="13" /></button>
            </div>
            <button class="add-row" @click="addVariable(envDraft.variables)"><Plus :size="13" /> {{ t.addVariable }}</button>
          </div>
          <div class="section-header narrow">
            <div><h2>Globals</h2><p>{{ t.localOnly }}</p></div>
            <button class="toolbar-btn" @click="saveGlobalsDraft"><Save :size="14" /> {{ t.save }}</button>
          </div>
          <div class="kv-table spacious">
            <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
            <div v-for="(variable, index) in globalsDraft" :key="variable.id" class="kv-row">
              <input v-model="variable.enabled" type="checkbox" />
              <input v-model="variable.key" />
              <input v-model="variable.value" />
              <input v-model="variable.description" />
              <button class="ghost-icon" @click="removeRow(globalsDraft, index)"><Trash2 :size="13" /></button>
            </div>
            <button class="add-row" @click="addVariable(globalsDraft)"><Plus :size="13" /> {{ t.addGlobal }}</button>
          </div>
        </template>

        <template v-else-if="activeNav === 'history'">
          <div class="section-header"><div><h2>{{ t.history }}</h2><p>SQLite</p></div></div>
          <div class="history-table">
            <div class="history-head"><span>METHOD</span><span>{{ t.response }}</span><span>{{ t.status }}</span><span>{{ t.duration }}</span></div>
            <button v-for="item in state?.history ?? []" :key="item.id" class="history-line" @click="selectRequest(item.request); activeNav = 'collections'">
              <span :class="['method', item.method.toLowerCase()]">{{ item.method }}</span>
              <span class="truncate">{{ item.url }}</span>
              <span :class="statusClass(item.statusCode)">{{ item.statusCode || '-' }}</span>
              <span>{{ item.durationMs }} ms</span>
            </button>
          </div>
        </template>

        <template v-else-if="activeNav === 'runner'">
          <div class="section-header">
            <div><h2>{{ t.runnerTitle }}</h2><p>{{ activeCollection?.name ?? t.collections }}</p></div>
            <button class="send-btn" :disabled="runnerBusy || !activeCollection" @click="runActiveCollection">
              <Loader2 v-if="runnerBusy" class="spin" :size="15" />
              <Play v-else :size="15" />
              {{ t.run }}
            </button>
          </div>
          <div v-if="runnerResult" class="runner-summary">
            <div><strong>{{ runnerResult.passed }}</strong><span>{{ t.passed }}</span></div>
            <div><strong>{{ runnerResult.failed }}</strong><span>{{ t.failed }}</span></div>
            <div><strong>{{ runnerResult.durationMs }} ms</strong><span>{{ t.duration }}</span></div>
          </div>
          <div class="response-panel standalone">
            <div v-for="item in runnerResult?.items ?? []" :key="item.name + item.message" class="test-row">
              <CheckCircle2 v-if="item.passed" :size="15" class="text-emerald-600" />
              <XCircle v-else :size="15" class="text-red-600" />
              <span>{{ item.name }}</span>
              <code v-if="item.message">{{ item.message }}</code>
            </div>
            <div v-if="!runnerResult" class="empty-panel">{{ t.runnerEmpty }}</div>
          </div>
        </template>

        <template v-else-if="activeNav === 'realtime'">
          <div class="section-header">
            <div><h2>{{ t.realtimeTitle }}</h2><p>{{ t.realtimeSubtitle }}</p></div>
          </div>
          <div class="realtime-grid">
            <section class="tool-panel">
              <div class="tool-panel-title">
                <div><strong>WebSocket</strong><span>{{ t.websocketDesc }}</span></div>
                <button class="send-btn" :disabled="realtimeBusy" @click="runWebSocketCheck">
                  <Loader2 v-if="realtimeBusy" class="spin" :size="15" />
                  <Send v-else :size="15" />
                  {{ t.connect }}
                </button>
              </div>
              <label class="stack-label inline"><span>URL</span><input v-model="wsDraft.url" class="field" placeholder="wss://echo.websocket.events" /></label>
              <label class="stack-label inline"><span>Message</span><textarea v-model="wsDraft.message" spellcheck="false" /></label>
              <div class="result-box">
                <div v-if="!wsResult" class="empty-panel">{{ t.wsEmpty }}</div>
                <template v-else>
                  <div class="kv-read"><span>{{ t.connected }}</span><code>{{ wsResult.connected ? 'yes' : 'no' }}</code></div>
                  <div class="kv-read"><span>{{ t.duration }}</span><code>{{ wsResult.durationMs }} ms</code></div>
                  <div v-if="wsResult.error" class="kv-read"><span>{{ t.error }}</span><code>{{ wsResult.error }}</code></div>
                  <div v-for="message in wsResult.received ?? []" :key="message" class="kv-read"><span>{{ t.received }}</span><code>{{ message }}</code></div>
                </template>
              </div>
            </section>

            <section class="tool-panel">
              <div class="tool-panel-title">
                <div><strong>SSE</strong><span>{{ t.sseDesc }}</span></div>
                <button class="send-btn" :disabled="realtimeBusy" @click="runSSECheck">
                  <Loader2 v-if="realtimeBusy" class="spin" :size="15" />
                  <Activity v-else :size="15" />
                  {{ t.listen }}
                </button>
              </div>
              <label class="stack-label inline"><span>URL</span><input v-model="sseDraft.url" class="field" placeholder="https://example.com/events" /></label>
              <div class="settings-grid compact-grid">
                <label><span>{{ t.maxEvents }}</span><input v-model.number="sseDraft.maxEvents" type="number" min="1" max="20" /></label>
                <label><span>{{ t.timeout }} (ms)</span><input v-model.number="sseDraft.timeoutMs" type="number" min="1000" step="1000" /></label>
              </div>
              <div class="result-box">
                <div v-if="!sseResult" class="empty-panel">{{ t.sseEmpty }}</div>
                <template v-else>
                  <div class="kv-read"><span>{{ t.status }}</span><code>{{ sseResult.statusCode || '-' }}</code></div>
                  <div class="kv-read"><span>{{ t.duration }}</span><code>{{ sseResult.durationMs }} ms</code></div>
                  <div v-if="sseResult.error" class="kv-read"><span>{{ t.error }}</span><code>{{ sseResult.error }}</code></div>
                  <div v-for="event in sseResult.events ?? []" :key="event" class="kv-read"><span>{{ t.event }}</span><code>{{ event }}</code></div>
                </template>
              </div>
            </section>
          </div>
        </template>

        <template v-else>
          <div class="section-header"><div><h2>{{ t.settings }}</h2><p>{{ t.settingsHelp }}</p></div></div>
          <div class="settings-grid">
            <label><span>{{ t.language }}</span><select v-model="language"><option value="zh-CN">中文</option><option value="en-US">English</option></select></label>
            <label><span>{{ t.theme }}</span><select v-model="settingsDraft.theme"><option value="light">{{ t.light }}</option></select></label>
            <div class="settings-note">
              <KeyRound :size="16" />
              {{ t.encryptedNote }}
            </div>
          </div>
        </template>

      </section>
    </main>

    <footer class="statusbar">
      <span>{{ statusMessage }}</span>
    </footer>
  </div>

  <Teleport to="body">
    <div v-if="activeModal" class="modal-backdrop" @click.self="closeModal">
      <section class="modal">
        <div class="modal-title">
          <strong>{{ activeModalTitle }}</strong>
          <button class="modal-close" type="button" title="Close" @click="closeModal">
            <X :size="14" :stroke-width="1.7" />
          </button>
        </div>
        <textarea v-if="activeModal === 'postman'" v-model="postmanText" spellcheck="false" :aria-label="t.postmanJSON"></textarea>
        <textarea v-else-if="activeModal === 'fetch'" v-model="fetchText" spellcheck="false" :aria-label="t.fetchSnippet"></textarea>
        <textarea v-else-if="activeModal === 'curl'" v-model="curlText" spellcheck="false" :aria-label="t.curlSnippet"></textarea>
        <textarea v-else v-model="exportText" spellcheck="false" readonly :aria-label="t.exportedCollection"></textarea>
        <button v-if="activeModal !== 'export'" class="send-btn" :disabled="busy" @click="activeModal === 'postman' ? importPostmanCollection() : activeModal === 'fetch' ? importFetchRequest() : importCurlRequest()">
          <Loader2 v-if="busy" class="spin" :size="15" />
          <Import v-else :size="15" />
          {{ activeModalTitle }}
        </button>
      </section>
    </div>
  </Teleport>
</template>
