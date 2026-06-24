<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { CheckCircle2, Clipboard, Clock3, Download, FileJson2, Loader2, Plus, RefreshCw, Search, Send, Trash2, Variable, Wand2, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import VoltSelect from './volt/VoltSelect.vue'
import JsonBodyEditor from './JsonBodyEditor.vue'
import VariableSuggestInput from './VariableSuggestInput.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltCheckbox from './volt/VoltCheckbox.vue'
import VoltInputText from './volt/VoltInputText.vue'
import VoltTabsBar from './volt/VoltTabsBar.vue'
import type { Translation } from '../i18n/messages'
import type { JsonPathOption, JsonToken, RequestTab, ResponseTab, ResponseView, VariableSuggestion } from '../types'
import { formatBytes, responseStatusText, statusClass } from '../utils/format'
import { tokenizeJSON } from '../utils/jsonHighlight'

const props = defineProps<{
  t: Translation
  activeCollection: domain.Collection | null
  response: domain.Response | null
  requestTabs: Array<{ key: RequestTab; label: string; count: number }>
  responseTabs: Array<{ key: ResponseTab; label: string; count: number }>
  highlightedResponseBody: JsonToken[]
  prettyResponseBody: string
  responseSearchMatches: number
  responseJSONPathResult: string
  responseJSONPathOptions: JsonPathOption[]
  busy: boolean
  methods: string[]
  authTypes: Array<{ value: string; label: string }>
  bodyModes: Array<{ value: string; label: string }>
  variableSuggestions: VariableSuggestion[]
  requestPreview: domain.PreparedRequest | null
  requestPreviewBusy: boolean
  variableDebugReport: domain.VariableDebugReport | null
  variableDebugBusy: boolean
  sendRequestAction: () => void | Promise<void>
}>()

const activeRequest = defineModel<domain.Request | null>('activeRequest', { required: true })
const activeRequestTab = defineModel<RequestTab>('activeRequestTab', { required: true })
const activeResponseTab = defineModel<ResponseTab>('activeResponseTab', { required: true })
const responseView = defineModel<ResponseView>('responseView', { required: true })
const responseSearch = defineModel<string>('responseSearch', { required: true })
const responseJSONPath = defineModel<string>('responseJSONPath', { required: true })
const responseVariableKey = defineModel<string>('responseVariableKey', { required: true })
const splitEditorRef = ref<HTMLElement | null>(null)
const requestTitleInput = ref<InstanceType<typeof VoltInputText> | null>(null)
const responseVariableKeyInput = ref<InstanceType<typeof VoltInputText> | null>(null)
const jsonPathOptionsRef = ref<HTMLElement | null>(null)

const emit = defineEmits<{
  'create-request': []
  'add-param': []
  'add-header': []
  'add-form-item': []
  'remove-row': [target: domain.KeyValue[], index: number]
  'remove-form-item': [index: number]
  'select-form-file': [index: number]
  'set-auth-type': [value: string]
  'query-json-path': []
  'copy-response-value': [value: string]
  'save-response': []
  'create-response-variable': []
  'refresh-request-preview': []
  'refresh-variable-debug': []
}>()

const formEditorMode = ref<'table' | 'text'>('table')
const responseHeightPercent = ref(Number(localStorage.getItem('restdeck.responseHeightPercent')) || 56)
const jsonPathSuggestionsOpen = ref(false)
const jsonPathSuggestionIndex = ref(0)
const responseVariableEditorOpen = ref(false)
const methodOptions = computed(() => props.methods.map((method) => ({ value: method, label: method })))
const addToOptions = computed(() => [{ value: 'header', label: props.t.headers }, { value: 'query', label: props.t.params }])
const formTypeOptions = computed(() => [{ value: 'text', label: props.t.text }, { value: 'file', label: props.t.file }])
const proxyModeOptions = computed(() => [
  { value: 'inherit', label: props.t.proxyInherit },
  { value: 'none', label: props.t.proxyNone },
  { value: 'custom', label: props.t.proxyCustom }
])
const editingRequestTitle = ref(false)
const responseInspectorText = computed({
  get: () => responseSearch.value,
  set: (value: string | number) => {
    const next = String(value ?? '')
    responseSearch.value = next
    responseJSONPath.value = isSupportedResponseJSONPath(next) ? next.trim() : ''
    jsonPathSuggestionsOpen.value = isJSONPathSuggestionInput(next)
    jsonPathSuggestionIndex.value = 0
  }
})
const responseInspectorQuery = computed(() => responseInspectorText.value.trim())
const responseSearchTerm = computed(() => isResponseJSONPathInput.value ? '' : responseInspectorQuery.value)
const isResponseJSONPathInput = computed(() => isSupportedResponseJSONPath(responseInspectorQuery.value))
const showJSONPathSuggestions = computed(() => jsonPathSuggestionsOpen.value && isJSONPathSuggestionInput(responseInspectorQuery.value))
const showJSONPathResultInContent = computed(() => isResponseJSONPathInput.value && props.responseJSONPathResult !== '')
const jsonPathResultBody = computed(() => {
  const raw = props.responseJSONPathResult
  if (!showJSONPathResultInContent.value || responseView.value !== 'pretty') return raw
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
})
const highlightedJSONPathResult = computed<JsonToken[]>(() => {
  if (responseView.value !== 'pretty') {
    return [{ type: 'plain', text: jsonPathResultBody.value }]
  }
  return tokenizeJSON(jsonPathResultBody.value)
})
const copyableResponseBody = computed(() => showJSONPathResultInContent.value ? jsonPathResultBody.value : props.prettyResponseBody)
const searchedResponseSegments = computed(() => {
  const query = responseSearchTerm.value
  const source = props.prettyResponseBody
  if (!query) return [{ text: source, match: false }]
  const lowerSource = source.toLowerCase()
  const lowerQuery = query.toLowerCase()
  const segments: Array<{ text: string; match: boolean }> = []
  let cursor = 0
  let index = lowerSource.indexOf(lowerQuery)
  while (index >= 0) {
    if (index > cursor) segments.push({ text: source.slice(cursor, index), match: false })
    segments.push({ text: source.slice(index, index + query.length), match: true })
    cursor = index + query.length
    index = lowerSource.indexOf(lowerQuery, cursor)
  }
  if (cursor < source.length) segments.push({ text: source.slice(cursor), match: false })
  return segments.length ? segments : [{ text: source, match: false }]
})

const visibleJSONPathOptions = computed(() => {
  if (!showJSONPathSuggestions.value) return []
  const query = responseInspectorQuery.value.toLowerCase()
  const options = props.responseJSONPathOptions
  if (query === '$' || query === '$.' || query === '$[') {
    return options.slice(0, 10)
  }
  const startsWith = options.filter((option) => option.path.toLowerCase().startsWith(query))
  const matched = startsWith.length ? startsWith : options.filter((option) => option.path.toLowerCase().includes(query))
  return matched.slice(0, 10)
})
const activeJSONPathSuggestionIndex = computed(() => {
  const length = visibleJSONPathOptions.value.length
  if (!length) return -1
  return Math.min(Math.max(jsonPathSuggestionIndex.value, 0), length - 1)
})
const activeResponseSearchMatches = computed(() => responseSearchTerm.value ? props.responseSearchMatches : 0)
const requestTitleInputStyle = computed(() => {
  const text = activeRequest.value?.name || props.t.requestName
  const units = Array.from(text).reduce((total, char) => total + (char.charCodeAt(0) > 255 ? 2 : 1), 0)
  return { width: `${Math.min(Math.max(units + 3, 12), 52)}ch` }
})
const splitEditorStyle = computed(() => ({
  '--response-height': `${responseHeightPercent.value}%`,
  '--request-height': `${100 - responseHeightPercent.value}%`
}))

let responseResizeStartY = 0
let responseResizeStartPercent = 56
let jsonPathSuggestionCloseTimer: ReturnType<typeof setTimeout> | null = null

function startResponseResize(event: PointerEvent) {
  const container = splitEditorRef.value
  if (!container) return
  event.preventDefault()
  responseResizeStartY = event.clientY
  responseResizeStartPercent = responseHeightPercent.value
  window.addEventListener('pointermove', resizeResponsePanel)
  window.addEventListener('pointerup', stopResponseResize, { once: true })
  document.body.classList.add('resizing-response-panel')
}

function resizeResponsePanel(event: PointerEvent) {
  const container = splitEditorRef.value
  if (!container) return
  const height = container.getBoundingClientRect().height
  if (height <= 0) return
  const deltaPercent = ((responseResizeStartY - event.clientY) / height) * 100
  const next = Math.min(78, Math.max(28, responseResizeStartPercent + deltaPercent))
  responseHeightPercent.value = Math.round(next * 10) / 10
}

function stopResponseResize() {
  window.removeEventListener('pointermove', resizeResponsePanel)
  localStorage.setItem('restdeck.responseHeightPercent', String(responseHeightPercent.value))
  document.body.classList.remove('resizing-response-panel')
}

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', resizeResponsePanel)
  window.removeEventListener('pointerup', stopResponseResize)
  if (jsonPathSuggestionCloseTimer) clearTimeout(jsonPathSuggestionCloseTimer)
  document.body.classList.remove('resizing-response-panel')
})

function editRequestTitle() {
  if (editingRequestTitle.value) {
    focusRequestTitle()
    return
  }
  editingRequestTitle.value = true
  focusRequestTitle()
}

function focusRequestTitle() {
  void nextTick(() => {
    requestTitleInput.value?.focus()
    requestTitleInput.value?.select()
  })
}

function finishRequestTitleEdit() {
  editingRequestTitle.value = false
}

function cancelRequestTitleEdit() {
  requestTitleInput.value?.blur()
  editingRequestTitle.value = false
}

watch(() => activeRequest.value?.id, () => {
  formEditorMode.value = 'table'
  editingRequestTitle.value = false
  ensureFormItems()
})

watch(() => activeRequest.value?.bodyMode, (mode) => {
  if (mode === 'form') {
    formEditorMode.value = 'table'
    ensureFormItems()
  }
})

watch(() => activeRequest.value?.body, (body) => {
  const request = activeRequest.value
  if (!request || request.bodyMode !== 'form' || formEditorMode.value !== 'text') return
  request.formItems = formItemsFromBody(body ?? '')
})

watch(isResponseJSONPathInput, (valid) => {
  if (valid) return
  responseVariableEditorOpen.value = false
  responseVariableKey.value = ''
})

watch(visibleJSONPathOptions, (options) => {
  if (!options.length) {
    jsonPathSuggestionIndex.value = 0
    return
  }
  if (jsonPathSuggestionIndex.value >= options.length) {
    jsonPathSuggestionIndex.value = options.length - 1
  }
})

function sendRequest() {
  void props.sendRequestAction()
}

function handleResponseInspectorKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown') {
    if (moveJSONPathSuggestion(1)) event.preventDefault()
    return
  }
  if (event.key === 'ArrowUp') {
    if (moveJSONPathSuggestion(-1)) event.preventDefault()
    return
  }
  if (event.key === 'Escape') {
    jsonPathSuggestionsOpen.value = false
    return
  }
  if (event.key !== 'Enter') return
  if (jsonPathSuggestionsOpen.value && activeJSONPathSuggestionIndex.value >= 0) {
    event.preventDefault()
    selectJSONPathOption(visibleJSONPathOptions.value[activeJSONPathSuggestionIndex.value], true)
    return
  }
  if (!isResponseJSONPathInput.value) return
  event.preventDefault()
  queryCurrentJSONPath()
}

function moveJSONPathSuggestion(delta: number) {
  if (!isJSONPathSuggestionInput(responseInspectorQuery.value)) return false
  jsonPathSuggestionsOpen.value = true
  const count = visibleJSONPathOptions.value.length
  if (!count) return false
  const current = activeJSONPathSuggestionIndex.value >= 0 ? activeJSONPathSuggestionIndex.value : 0
  jsonPathSuggestionIndex.value = (current + delta + count) % count
  void nextTick(scrollActiveJSONPathSuggestionIntoView)
  return true
}

function scrollActiveJSONPathSuggestionIntoView() {
  const active = jsonPathOptionsRef.value?.querySelector('button.active')
  if (active instanceof HTMLElement) {
    active.scrollIntoView({ block: 'nearest' })
  }
}

function selectJSONPathOption(option: JsonPathOption, query = false) {
  responseInspectorText.value = option.path
  jsonPathSuggestionsOpen.value = false
  if (query) queryCurrentJSONPath()
}

function queryCurrentJSONPath() {
  jsonPathSuggestionsOpen.value = false
  emit('query-json-path')
}

function openJSONPathSuggestions() {
  if (jsonPathSuggestionCloseTimer) clearTimeout(jsonPathSuggestionCloseTimer)
  jsonPathSuggestionsOpen.value = isJSONPathSuggestionInput(responseInspectorQuery.value)
}

function closeJSONPathSuggestionsSoon() {
  if (jsonPathSuggestionCloseTimer) clearTimeout(jsonPathSuggestionCloseTimer)
  jsonPathSuggestionCloseTimer = setTimeout(() => {
    jsonPathSuggestionsOpen.value = false
  }, 120)
}

function openResponseVariableEditor() {
  if (!isResponseJSONPathInput.value) return
  responseVariableEditorOpen.value = !responseVariableEditorOpen.value
  void nextTick(() => {
    responseVariableKeyInput.value?.focus()
  })
}

function confirmResponseVariable() {
  if (!isResponseJSONPathInput.value) return
  emit('create-response-variable')
  responseVariableEditorOpen.value = false
}

function isJSONPathSuggestionInput(value: string) {
  const input = value.trim()
  return input === '$' || input.startsWith('$.') || input.startsWith('$[')
}

function isSupportedResponseJSONPath(value: string) {
  const input = value.trim()
  if (input === '$') return true
  if (!input.startsWith('$.') && !input.startsWith('$[')) return false
  return parseSupportedJSONPath(input)
}

function parseSupportedJSONPath(path: string) {
  let rest = path.slice(1)
  while (rest) {
    if (rest.startsWith('.')) {
      rest = rest.slice(1)
      const match = rest.match(/^[^.[\]]+/)
      if (!match) return false
      rest = rest.slice(match[0].length)
      continue
    }
    if (rest.startsWith('[')) {
      const end = rest.indexOf(']')
      if (end < 0) return false
      const token = rest.slice(1, end).trim()
      if (!/^\d+$/.test(token) && !/^"([^"\\]|\\.)*"$/.test(token) && !/^'([^'\\]|\\.)*'$/.test(token)) {
        return false
      }
      rest = rest.slice(end + 1)
      continue
    }
    return false
  }
  return true
}

function formatRequestJSON() {
  if (!activeRequest.value) return
  const raw = activeRequest.value.body?.trim()
  if (!raw) return
  try {
    activeRequest.value.body = JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    // Keep invalid JSON untouched while the user is still editing.
  }
}

function switchFormEditorMode(mode: 'table' | 'text') {
  if (!activeRequest.value) return
  if (mode === 'text') {
    activeRequest.value.body = formItemsToBody(activeRequest.value.formItems ?? [])
  } else {
    activeRequest.value.formItems = formItemsFromBody(activeRequest.value.body ?? '')
  }
  formEditorMode.value = mode
}

function toggleFormEditorMode() {
  switchFormEditorMode(formEditorMode.value === 'table' ? 'text' : 'table')
}

function ensureFormItems() {
  const request = activeRequest.value
  if (!request || request.bodyMode !== 'form') return
  if (!request.formItems?.length) {
    request.formItems = formItemsFromBody(request.body ?? '')
  }
  if (!request.formItems.length) {
    request.formItems.push(newFormItem())
  }
}

function setFormItemType(item: domain.FormItem, type: string) {
  item.type = type === 'file' ? 'file' : 'text'
  if (item.type === 'file') {
    item.value = ''
  } else {
    item.filePath = ''
  }
}

function formItemsToBody(items: domain.FormItem[]) {
  return (items ?? [])
    .filter((item) => item.key || item.value || item.filePath)
    .map((item) => `${item.key}=${item.type === 'file' ? `@${item.filePath}` : item.value}`)
    .join('\n')
}

function formItemsFromBody(raw: string) {
  const items = raw
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
  return items.length ? items : [newFormItem()]
}

function newFormItem() {
  return new domain.FormItem({ id: crypto.randomUUID(), enabled: true, key: '', type: 'text', value: '', filePath: '', description: '' })
}
</script>

<template>
  <template v-if="activeRequest">
    <div class="editor-header">
      <div class="breadcrumb">
        <span>{{ activeCollection?.name ?? 'Collection' }}</span>
        <span>/</span>
        <VoltInputText
          v-if="editingRequestTitle"
          ref="requestTitleInput"
          v-model="activeRequest.name"
          input-class="title-input"
          :input-style="requestTitleInputStyle"
          @blur="finishRequestTitleEdit"
          @keydown.enter.prevent="requestTitleInput?.blur()"
          @keydown.esc.prevent="cancelRequestTitleEdit"
        />
        <button
          v-else
          class="title-display"
          type="button"
          @click="editRequestTitle"
        >
          {{ activeRequest.name || t.requestName }}
        </button>
      </div>
    </div>

    <div class="request-line">
      <VoltSelect v-model="activeRequest.method" class="method-select" :options="methodOptions" />
      <VariableSuggestInput v-model="activeRequest.url" input-class="url-input" :suggestions="variableSuggestions" placeholder="https://api.example.com/v1/resource" />
      <VoltButton class="send-btn" :disabled="busy" @click="sendRequest">
        <Loader2 v-if="busy" class="spin" :size="15" />
        <Send v-else :size="15" />
        {{ t.send }}
      </VoltButton>
    </div>

    <div ref="splitEditorRef" class="split-editor" :style="splitEditorStyle">
      <section class="request-editor">
        <VoltTabsBar v-model="activeRequestTab" :items="requestTabs" />

        <div class="tab-panel">
          <div v-if="activeRequestTab === 'params'" class="kv-table">
            <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
            <div v-for="(param, index) in activeRequest.params" :key="param.id" class="kv-row">
              <VoltCheckbox v-model="param.enabled" />
              <VoltInputText v-model="param.key" :placeholder="t.key" />
              <VariableSuggestInput v-model="param.value" :suggestions="variableSuggestions" :placeholder="t.value" />
              <VoltInputText v-model="param.description" :placeholder="t.description" />
              <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('remove-row', activeRequest!.params, index)"><Trash2 :size="13" /></VoltButton>
            </div>
            <VoltButton class="add-row" variant="secondary" @click="emit('add-param')"><Plus :size="13" /> {{ t.addParam }}</VoltButton>
          </div>

          <div v-else-if="activeRequestTab === 'headers'" class="kv-table">
            <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
            <div v-for="(header, index) in activeRequest.headers" :key="header.id" class="kv-row">
              <VoltCheckbox v-model="header.enabled" />
              <VoltInputText v-model="header.key" :placeholder="t.headers" />
              <VariableSuggestInput v-model="header.value" :suggestions="variableSuggestions" :placeholder="t.value" />
              <VoltInputText v-model="header.description" :placeholder="t.description" />
              <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('remove-row', activeRequest!.headers, index)"><Trash2 :size="13" /></VoltButton>
            </div>
            <VoltButton class="add-row" variant="secondary" @click="emit('add-header')"><Plus :size="13" /> {{ t.addHeader }}</VoltButton>
          </div>

          <div v-else-if="activeRequestTab === 'auth'" class="auth-grid">
            <label>
              <span>{{ t.type }}</span>
              <VoltSelect :model-value="activeRequest.auth?.type ?? 'none'" :options="authTypes" @change="emit('set-auth-type', String($event))" />
            </label>
            <template v-if="activeRequest.auth?.type === 'apiKey'">
              <label><span>{{ t.key }}</span><VariableSuggestInput v-model="activeRequest.auth.values.key" :suggestions="variableSuggestions" /></label>
              <label><span>{{ t.value }}</span><VariableSuggestInput v-model="activeRequest.auth.values.value" type="password" :suggestions="variableSuggestions" /></label>
              <label><span>{{ t.addTo }}</span><VoltSelect v-model="activeRequest.auth.values.in" :options="addToOptions" /></label>
            </template>
            <template v-else-if="activeRequest.auth?.type === 'bearer'">
              <label><span>Token</span><VariableSuggestInput v-model="activeRequest.auth.values.token" type="password" :suggestions="variableSuggestions" /></label>
            </template>
            <template v-else-if="activeRequest.auth?.type === 'basic' || activeRequest.auth?.type === 'digest'">
              <label><span>{{ t.username }}</span><VariableSuggestInput v-model="activeRequest.auth.values.username" :suggestions="variableSuggestions" /></label>
              <label><span>{{ t.password }}</span><VariableSuggestInput v-model="activeRequest.auth.values.password" type="password" :suggestions="variableSuggestions" /></label>
            </template>
            <template v-else-if="activeRequest.auth?.type === 'oauth1'">
              <label><span>{{ t.consumerKey }}</span><VariableSuggestInput v-model="activeRequest.auth.values.consumerKey" :suggestions="variableSuggestions" /></label>
              <label><span>{{ t.consumerSecret }}</span><VariableSuggestInput v-model="activeRequest.auth.values.consumerSecret" type="password" :suggestions="variableSuggestions" /></label>
              <label><span>Token</span><VariableSuggestInput v-model="activeRequest.auth.values.token" :suggestions="variableSuggestions" /></label>
              <label><span>{{ t.tokenSecret }}</span><VariableSuggestInput v-model="activeRequest.auth.values.tokenSecret" type="password" :suggestions="variableSuggestions" /></label>
            </template>
            <template v-else-if="activeRequest.auth?.type === 'oauth2'">
              <label><span>{{ t.accessToken }}</span><VariableSuggestInput v-model="activeRequest.auth.values.accessToken" type="password" :suggestions="variableSuggestions" /></label>
            </template>
            <p v-else class="muted">{{ t.noAuth }}</p>
          </div>

          <div v-else-if="activeRequestTab === 'body'" class="body-editor">
            <div class="body-toolbar">
              <VoltSelect v-model="activeRequest.bodyMode" class="field compact" :options="bodyModes" />
              <VoltButton v-if="activeRequest.bodyMode === 'json'" class="toolbar-btn" @click="formatRequestJSON">
                <Wand2 :size="14" />
                {{ t.formatJSON }}
              </VoltButton>
              <VoltButton v-if="activeRequest.bodyMode === 'form'" class="toolbar-btn" @click="toggleFormEditorMode">
                {{ formEditorMode === 'table' ? t.formViewMode1 : t.formViewMode2 }}
              </VoltButton>
            </div>
            <JsonBodyEditor v-if="activeRequest.bodyMode === 'json'" v-model="activeRequest.body" :suggestions="variableSuggestions" />
            <template v-else-if="activeRequest.bodyMode === 'form'">
              <div v-if="formEditorMode === 'table'" class="kv-table form-table">
                <div class="kv-head form-head"><span></span><span>{{ t.key }}</span><span>{{ t.type }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
                <div v-for="(item, index) in activeRequest.formItems" :key="item.id" class="kv-row form-row">
                  <VoltCheckbox v-model="item.enabled" />
                  <VoltInputText v-model="item.key" :placeholder="t.key" />
                  <VoltSelect :model-value="item.type" :options="formTypeOptions" @change="setFormItemType(item, String($event))" />
                  <div class="form-value-cell">
                    <VariableSuggestInput v-if="item.type !== 'file'" v-model="item.value" :suggestions="variableSuggestions" :placeholder="t.value" />
                    <template v-else>
                      <VoltButton class="small-btn" size="sm" @click="emit('select-form-file', index)">{{ t.chooseFile }}</VoltButton>
                      <span class="file-path" :title="item.filePath">{{ item.filePath }}</span>
                    </template>
                  </div>
                  <VoltInputText v-model="item.description" :placeholder="t.description" />
                  <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('remove-form-item', index)"><Trash2 :size="13" /></VoltButton>
                </div>
                <VoltButton class="add-row" variant="secondary" @click="emit('add-form-item')"><Plus :size="13" /> {{ t.addFormItem }}</VoltButton>
              </div>
              <VariableSuggestInput v-else v-model="activeRequest.body" as="textarea" :suggestions="variableSuggestions" :spellcheck="false" placeholder="name=value&#10;avatar=@D:\path\file.png" />
            </template>
            <VariableSuggestInput v-else-if="activeRequest.bodyMode !== 'none'" v-model="activeRequest.body" as="textarea" :suggestions="variableSuggestions" :spellcheck="false" placeholder='{"hello": "world"}' />
            <div v-else class="empty-panel">{{ t.noBody }}</div>
          </div>

          <div v-else-if="activeRequestTab === 'pre'" class="body-editor">
            <VariableSuggestInput v-model="activeRequest.preScript" as="textarea" :suggestions="variableSuggestions" :spellcheck="false" placeholder="pm.variables.set('traceId', '{{$guid}}');" />
          </div>

          <div v-else-if="activeRequestTab === 'tests'" class="body-editor">
            <VariableSuggestInput v-model="activeRequest.testScript" as="textarea" :suggestions="variableSuggestions" :spellcheck="false" placeholder='pm.test("Status is 200", function () { expect(pm.response.code).to.equal(200); });' />
          </div>

          <div v-else-if="activeRequestTab === 'preview'" class="debug-panel">
            <div class="debug-toolbar">
              <strong>{{ t.actualRequest }}</strong>
              <VoltButton variant="secondary" :disabled="requestPreviewBusy" @click="emit('refresh-request-preview')">
                <Loader2 v-if="requestPreviewBusy" class="spin" :size="14" />
                <RefreshCw v-else :size="14" />
                {{ t.refresh }}
              </VoltButton>
            </div>
            <div v-if="requestPreview?.error" class="debug-error">{{ requestPreview.error }}</div>
            <div v-if="requestPreview" class="debug-grid">
              <span>{{ t.method }}</span><code>{{ requestPreview.method || '-' }}</code>
              <span>{{ t.requestUrl }}</span><code>{{ requestPreview.url || '-' }}</code>
              <span>{{ t.proxyMode }}</span><code>{{ requestPreview.proxyApplied ? `${requestPreview.proxy?.url || '-'} (${requestPreview.proxySource})` : (requestPreview.proxyExcluded ? t.proxyExcluded : t.proxyNone) }}</code>
              <span>{{ t.body }}</span><code>{{ requestPreview.body?.contentType || requestPreview.body?.mode || '-' }} · {{ requestPreview.body?.sizeBytes ?? 0 }} bytes</code>
            </div>
            <div v-if="requestPreview?.headers?.length" class="debug-section">
              <strong>{{ t.headers }}</strong>
              <div class="kv-read" v-for="header in requestPreview.headers" :key="header.key"><span>{{ header.key }}</span><code>{{ header.value }}</code></div>
            </div>
            <div v-if="requestPreview?.cookies?.length" class="debug-section">
              <strong>{{ t.cookies }}</strong>
              <div class="kv-read" v-for="cookie in requestPreview.cookies" :key="`${cookie.domain}-${cookie.name}`"><span>{{ cookie.name }}</span><code>{{ cookie.value }}</code></div>
            </div>
            <div v-if="requestPreview?.body?.text" class="debug-section">
              <strong>{{ t.body }}</strong>
              <pre>{{ requestPreview.body.text }}</pre>
            </div>
            <div v-if="!requestPreview" class="empty-panel">{{ t.actualRequestEmpty }}</div>
          </div>

          <div v-else-if="activeRequestTab === 'variables'" class="debug-panel">
            <div class="debug-toolbar">
              <strong>{{ t.variableDebug }}</strong>
              <VoltButton variant="secondary" :disabled="variableDebugBusy" @click="emit('refresh-variable-debug')">
                <Loader2 v-if="variableDebugBusy" class="spin" :size="14" />
                <RefreshCw v-else :size="14" />
                {{ t.refresh }}
              </VoltButton>
            </div>
            <div v-for="error in variableDebugReport?.errors ?? []" :key="error" class="debug-error">{{ error }}</div>
            <div v-if="variableDebugReport?.variables?.length" class="kv-table variable-debug-table">
              <div class="kv-head variable-debug-head"><span>{{ t.key }}</span><span>{{ t.type }}</span><span>{{ t.value }}</span><span>{{ t.result }}</span></div>
              <div v-for="variable in variableDebugReport.variables" :key="`${variable.source}-${variable.name}`" class="kv-row variable-debug-row">
                <code>{{ variable.name }}</code>
                <span>{{ variable.type }}</span>
                <code>{{ variable.raw || '-' }}</code>
                <code :class="{ 'text-red-600': variable.error }">{{ variable.error || variable.value || '-' }}</code>
              </div>
            </div>
            <div v-else class="empty-panel">{{ t.variableDebugEmpty }}</div>
          </div>

          <div v-else-if="activeRequestTab === 'settings'" class="settings-sections">
            <section class="settings-group">
              <div class="settings-group-title">{{ t.requestSettings }}</div>
              <div class="settings-fields">
                <label class="settings-field">
                  <span>{{ t.timeout }} (ms)</span>
                  <VoltInputText v-model="activeRequest.timeoutMs" type="number" />
                </label>
              </div>
            </section>
            <section class="settings-group">
              <div class="settings-group-title">{{ t.proxySettings }}</div>
              <div class="settings-fields">
                <label class="settings-field">
                  <span>{{ t.proxyMode }}</span>
                  <VoltSelect v-model="activeRequest.proxy.mode" :options="proxyModeOptions" />
                </label>
                <label v-if="activeRequest.proxy.mode === 'custom'" class="settings-field">
                  <span>{{ t.proxyUrl }}</span>
                  <VariableSuggestInput v-model="activeRequest.proxy.url" :suggestions="variableSuggestions" placeholder="http://127.0.0.1:7890" />
                </label>
              </div>
            </section>
          </div>
        </div>
      </section>

      <div class="response-resizer" role="separator" aria-orientation="horizontal" @pointerdown="startResponseResize"></div>

      <section class="response-editor">
        <div class="response-meta">
          <strong>{{ t.response }}</strong>
          <span v-if="response" :class="statusClass(response.statusCode)">{{ responseStatusText(response) }}</span>
          <span v-if="response"><Clock3 :size="13" /> {{ response.durationMs }} ms</span>
          <span v-if="response">{{ formatBytes(response.sizeBytes) }}</span>
        </div>
        <VoltTabsBar v-model="activeResponseTab" :items="responseTabs" />
        <div class="response-panel" :class="{ 'with-body-tools': activeResponseTab === 'body' && response }">
          <template v-if="!response">
            <div class="empty-panel">{{ t.noResponse }}</div>
          </template>
          <template v-else-if="activeResponseTab === 'body'">
            <div class="response-tools">
              <div class="response-toolbar">
                <div class="view-switch">
                  <VoltButton :class="{ active: responseView === 'pretty' }" @click="responseView = 'pretty'">{{ t.pretty }}</VoltButton>
                  <VoltButton :class="{ active: responseView === 'raw' }" @click="responseView = 'raw'">{{ t.raw }}</VoltButton>
                  <VoltButton :class="{ active: responseView === 'preview' }" @click="responseView = 'preview'">{{ t.preview }}</VoltButton>
                  <span class="pill">JSON</span>
                </div>
                <div class="response-search-box">
                  <Search :size="13" />
                  <VoltInputText
                    v-model="responseInspectorText"
                    :placeholder="t.responseInspectorPlaceholder"
                    @focus="openJSONPathSuggestions"
                    @blur="closeJSONPathSuggestionsSoon"
                    @keydown="handleResponseInspectorKeydown"
                  />
                  <span>{{ activeResponseSearchMatches }}</span>
                  <div v-if="visibleJSONPathOptions.length" ref="jsonPathOptionsRef" class="jsonpath-options">
                    <button
                      v-for="(option, index) in visibleJSONPathOptions"
                      :key="option.path"
                      type="button"
                      :class="{ active: index === activeJSONPathSuggestionIndex }"
                      @mousedown.prevent="selectJSONPathOption(option, true)"
                    >
                      <code>{{ option.path }}</code>
                      <span>{{ option.preview }}</span>
                    </button>
                  </div>
                </div>
                <VoltButton variant="secondary" @click="emit('copy-response-value', copyableResponseBody)"><Clipboard :size="14" /> {{ t.copyResult }}</VoltButton>
                <div class="response-variable-creator">
                  <VoltButton variant="secondary" :disabled="!isResponseJSONPathInput" @click="openResponseVariableEditor"><Variable :size="14" /> {{ t.createVariableFromResponse }}</VoltButton>
                  <div v-if="responseVariableEditorOpen" class="response-variable-popover">
                    <strong>{{ t.createVariableFromResponse }}</strong>
                    <div class="response-variable-popover-body">
                      <VoltInputText ref="responseVariableKeyInput" v-model="responseVariableKey" class="response-variable-input" :placeholder="t.key" @keydown.enter.prevent="confirmResponseVariable" />
                      <VoltButton class="response-variable-confirm" variant="primary" size="sm" @click="confirmResponseVariable">{{ t.confirm }}</VoltButton>
                    </div>
                  </div>
                </div>
                <VoltButton variant="secondary" @click="emit('save-response')"><Download :size="14" /> {{ t.saveResponse }}</VoltButton>
              </div>
            </div>
            <div class="response-content">
              <iframe v-if="showJSONPathResultInContent && responseView === 'preview'" :srcdoc="jsonPathResultBody" />
              <pre v-else-if="showJSONPathResultInContent && responseView === 'pretty'" class="json-highlight"><span v-for="(token, index) in highlightedJSONPathResult" :key="index" :class="`json-${token.type}`">{{ token.text }}</span></pre>
              <pre v-else-if="showJSONPathResultInContent">{{ jsonPathResultBody }}</pre>
              <iframe v-else-if="responseView === 'preview'" :srcdoc="response.body" />
              <pre v-else-if="responseSearchTerm" class="json-highlight"><span v-for="(segment, index) in searchedResponseSegments" :key="index" :class="{ 'search-match': segment.match }">{{ segment.text }}</span></pre>
              <pre v-else-if="responseView === 'pretty'" class="json-highlight"><span v-for="(token, index) in highlightedResponseBody" :key="index" :class="`json-${token.type}`">{{ token.text }}</span></pre>
              <pre v-else>{{ prettyResponseBody }}</pre>
            </div>
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
  </template>

  <div v-else class="blank-state">
    <FileJson2 :size="28" />
    <span>{{ t.createOrSelect }}</span>
    <VoltButton class="send-btn" @click="emit('create-request')"><Plus :size="15" /> {{ t.newRequest }}</VoltButton>
  </div>
</template>
