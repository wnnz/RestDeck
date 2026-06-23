<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { CheckCircle2, Clock3, Download, FileJson2, Loader2, Plus, Save, Send, Trash2, Wand2, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import VoltSelect from './volt/VoltSelect.vue'
import JsonBodyEditor from './JsonBodyEditor.vue'
import VariableSuggestInput from './VariableSuggestInput.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltCheckbox from './volt/VoltCheckbox.vue'
import VoltInputText from './volt/VoltInputText.vue'
import VoltTabsBar from './volt/VoltTabsBar.vue'
import type { Translation } from '../i18n/messages'
import type { JsonToken, RequestTab, ResponseTab, ResponseView, VariableSuggestion } from '../types'
import { formatBytes, responseStatusText, statusClass } from '../utils/format'

const props = defineProps<{
  t: Translation
  activeCollection: domain.Collection | null
  response: domain.Response | null
  requestTabs: Array<{ key: RequestTab; label: string; count: number }>
  responseTabs: Array<{ key: ResponseTab; label: string; count: number }>
  highlightedResponseBody: JsonToken[]
  prettyResponseBody: string
  busy: boolean
  methods: string[]
  authTypes: Array<{ value: string; label: string }>
  bodyModes: Array<{ value: string; label: string }>
  variableSuggestions: VariableSuggestion[]
  sendRequestAction: () => void | Promise<void>
}>()

const activeRequest = defineModel<domain.Request | null>('activeRequest', { required: true })
const activeRequestTab = defineModel<RequestTab>('activeRequestTab', { required: true })
const activeResponseTab = defineModel<ResponseTab>('activeResponseTab', { required: true })
const responseView = defineModel<ResponseView>('responseView', { required: true })

const emit = defineEmits<{
  'save-request': []
  'delete-request': []
  'export-collection': []
  'create-request': []
  'add-param': []
  'add-header': []
  'add-form-item': []
  'remove-row': [target: domain.KeyValue[], index: number]
  'remove-form-item': [index: number]
  'select-form-file': [index: number]
  'set-auth-type': [value: string]
}>()

const formEditorMode = ref<'table' | 'text'>('table')
const methodOptions = computed(() => props.methods.map((method) => ({ value: method, label: method })))
const addToOptions = computed(() => [{ value: 'header', label: props.t.headers }, { value: 'query', label: props.t.params }])
const formTypeOptions = computed(() => [{ value: 'text', label: props.t.text }, { value: 'file', label: props.t.file }])
const proxyModeOptions = computed(() => [
  { value: 'inherit', label: props.t.proxyInherit },
  { value: 'none', label: props.t.proxyNone },
  { value: 'custom', label: props.t.proxyCustom }
])

watch(() => activeRequest.value?.id, () => {
  formEditorMode.value = 'table'
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

function sendRequest() {
  void props.sendRequestAction()
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
        <VoltInputText v-model="activeRequest.name" input-class="title-input" />
      </div>
      <div class="editor-actions">
        <VoltButton class="toolbar-btn" :disabled="busy" @click="emit('save-request')">
          <Save :size="14" />
          {{ t.save }}
        </VoltButton>
        <VoltButton class="toolbar-btn" :disabled="!activeRequest.id || busy" @click="emit('delete-request')">
          <Trash2 :size="14" />
          {{ t.delete }}
        </VoltButton>
        <VoltButton class="toolbar-btn" :disabled="!activeCollection" @click="emit('export-collection')">
          <Download :size="14" />
          {{ t.export }}
        </VoltButton>
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

    <div class="split-editor">
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

          <div v-else class="settings-sections">
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

      <section class="response-editor">
        <div class="response-meta">
          <strong>{{ t.response }}</strong>
          <span v-if="response" :class="statusClass(response.statusCode)">{{ responseStatusText(response) }}</span>
          <span v-if="response"><Clock3 :size="13" /> {{ response.durationMs }} ms</span>
          <span v-if="response">{{ formatBytes(response.sizeBytes) }}</span>
        </div>
        <VoltTabsBar v-model="activeResponseTab" :items="responseTabs" />
        <div class="response-panel">
          <template v-if="!response">
            <div class="empty-panel">{{ t.noResponse }}</div>
          </template>
          <template v-else-if="activeResponseTab === 'body'">
            <div class="view-switch">
              <VoltButton :class="{ active: responseView === 'pretty' }" @click="responseView = 'pretty'">{{ t.pretty }}</VoltButton>
              <VoltButton :class="{ active: responseView === 'raw' }" @click="responseView = 'raw'">{{ t.raw }}</VoltButton>
              <VoltButton :class="{ active: responseView === 'preview' }" @click="responseView = 'preview'">{{ t.preview }}</VoltButton>
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
  </template>

  <div v-else class="blank-state">
    <FileJson2 :size="28" />
    <span>{{ t.createOrSelect }}</span>
    <VoltButton class="send-btn" @click="emit('create-request')"><Plus :size="15" /> New request</VoltButton>
  </div>
</template>
