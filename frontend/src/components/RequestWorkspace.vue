<script setup lang="ts">
import { ref, watch } from 'vue'
import { CheckCircle2, Clock3, Download, FileJson2, Loader2, Plus, Save, Send, Trash2, Wand2, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import JsonBodyEditor from './JsonBodyEditor.vue'
import type { Translation } from '../i18n/messages'
import type { JsonToken, RequestTab, ResponseTab, ResponseView } from '../types'
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
        <input v-model="activeRequest.name" class="title-input" />
      </div>
      <div class="editor-actions">
        <button class="toolbar-btn" :disabled="busy" @click="emit('save-request')">
          <Save :size="14" />
          {{ t.save }}
        </button>
        <button class="toolbar-btn" :disabled="!activeRequest.id || busy" @click="emit('delete-request')">
          <Trash2 :size="14" />
          {{ t.delete }}
        </button>
        <button class="toolbar-btn" :disabled="!activeCollection" @click="emit('export-collection')">
          <Download :size="14" />
          {{ t.export }}
        </button>
      </div>
    </div>

    <div class="request-line">
      <select v-model="activeRequest.method" class="method-select">
        <option v-for="method in methods" :key="method" :value="method">{{ method }}</option>
      </select>
      <input v-model="activeRequest.url" class="url-input" placeholder="https://api.example.com/v1/resource" />
      <button class="send-btn" :disabled="busy" @click="sendRequest">
        <Loader2 v-if="busy" class="spin" :size="15" />
        <Send v-else :size="15" />
        {{ t.send }}
      </button>
    </div>

    <div class="split-editor">
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
              <button class="ghost-icon" @click="emit('remove-row', activeRequest!.params, index)"><Trash2 :size="13" /></button>
            </div>
            <button class="add-row" @click="emit('add-param')"><Plus :size="13" /> {{ t.addParam }}</button>
          </div>

          <div v-else-if="activeRequestTab === 'headers'" class="kv-table">
            <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
            <div v-for="(header, index) in activeRequest.headers" :key="header.id" class="kv-row">
              <input v-model="header.enabled" type="checkbox" />
              <input v-model="header.key" :placeholder="t.headers" />
              <input v-model="header.value" :placeholder="t.value" />
              <input v-model="header.description" :placeholder="t.description" />
              <button class="ghost-icon" @click="emit('remove-row', activeRequest!.headers, index)"><Trash2 :size="13" /></button>
            </div>
            <button class="add-row" @click="emit('add-header')"><Plus :size="13" /> {{ t.addHeader }}</button>
          </div>

          <div v-else-if="activeRequestTab === 'auth'" class="auth-grid">
            <label>
              <span>{{ t.type }}</span>
              <select :value="activeRequest.auth?.type ?? 'none'" @change="emit('set-auth-type', ($event.target as HTMLSelectElement).value)">
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
            <div class="body-toolbar">
              <select v-model="activeRequest.bodyMode" class="field compact">
                <option v-for="mode in bodyModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
              </select>
              <button v-if="activeRequest.bodyMode === 'json'" class="toolbar-btn" type="button" @click="formatRequestJSON">
                <Wand2 :size="14" />
                {{ t.formatJSON }}
              </button>
              <button v-if="activeRequest.bodyMode === 'form'" class="toolbar-btn" type="button" @click="toggleFormEditorMode">
                {{ formEditorMode === 'table' ? t.formViewMode1 : t.formViewMode2 }}
              </button>
            </div>
            <JsonBodyEditor v-if="activeRequest.bodyMode === 'json'" v-model="activeRequest.body" />
            <template v-else-if="activeRequest.bodyMode === 'form'">
              <div v-if="formEditorMode === 'table'" class="kv-table form-table">
                <div class="kv-head form-head"><span></span><span>{{ t.key }}</span><span>{{ t.type }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
                <div v-for="(item, index) in activeRequest.formItems" :key="item.id" class="kv-row form-row">
                  <input v-model="item.enabled" type="checkbox" />
                  <input v-model="item.key" :placeholder="t.key" />
                  <select :value="item.type" @change="setFormItemType(item, ($event.target as HTMLSelectElement).value)">
                    <option value="text">{{ t.text }}</option>
                    <option value="file">{{ t.file }}</option>
                  </select>
                  <div class="form-value-cell">
                    <input v-if="item.type !== 'file'" v-model="item.value" :placeholder="t.value" />
                    <template v-else>
                      <button class="small-btn" type="button" @click="emit('select-form-file', index)">{{ t.chooseFile }}</button>
                      <span class="file-path" :title="item.filePath">{{ item.filePath }}</span>
                    </template>
                  </div>
                  <input v-model="item.description" :placeholder="t.description" />
                  <button class="ghost-icon" @click="emit('remove-form-item', index)"><Trash2 :size="13" /></button>
                </div>
                <button class="add-row" @click="emit('add-form-item')"><Plus :size="13" /> {{ t.addFormItem }}</button>
              </div>
              <textarea v-else v-model="activeRequest.body" spellcheck="false" placeholder="name=value&#10;avatar=@D:\path\file.png"></textarea>
            </template>
            <textarea v-else-if="activeRequest.bodyMode !== 'none'" v-model="activeRequest.body" spellcheck="false" placeholder='{"hello": "world"}'></textarea>
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
  </template>

  <div v-else class="blank-state">
    <FileJson2 :size="28" />
    <span>{{ t.createOrSelect }}</span>
    <button class="send-btn" @click="emit('create-request')"><Plus :size="15" /> New request</button>
  </div>
</template>
