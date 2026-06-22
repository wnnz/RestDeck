<script setup lang="ts">
import { Activity, Loader2, Send } from 'lucide-vue-next'
import { domain, realtime } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { VariableSuggestion } from '../types'
import VariableSuggestInput from './VariableSuggestInput.vue'

defineProps<{
  t: Translation
  variableSuggestions: VariableSuggestion[]
  realtimeBusy: boolean
  wsResult: realtime.WebSocketResult | null
  sseResult: realtime.SSEResult | null
}>()

const wsDraft = defineModel<{ url: string; message: string; headers: domain.KeyValue[]; proxy: domain.ProxyConfig; timeoutMs: number }>('wsDraft', { required: true })
const sseDraft = defineModel<{ url: string; headers: domain.KeyValue[]; proxy: domain.ProxyConfig; timeoutMs: number; maxEvents: number }>('sseDraft', { required: true })

const emit = defineEmits<{
  runWebSocket: []
  runSSE: []
}>()
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.realtimeTitle }}</h2><p>{{ t.realtimeSubtitle }}</p></div>
  </div>
  <div class="realtime-grid">
    <section class="tool-panel">
      <div class="tool-panel-title">
        <div><strong>WebSocket</strong><span>{{ t.websocketDesc }}</span></div>
        <button class="send-btn" :disabled="realtimeBusy" @click="emit('runWebSocket')">
          <Loader2 v-if="realtimeBusy" class="spin" :size="15" />
          <Send v-else :size="15" />
          {{ t.connect }}
        </button>
      </div>
      <label class="stack-label inline"><span>URL</span><VariableSuggestInput v-model="wsDraft.url" input-class="field" :suggestions="variableSuggestions" placeholder="wss://echo.websocket.events" /></label>
      <label class="stack-label inline"><span>Message</span><VariableSuggestInput v-model="wsDraft.message" as="textarea" :suggestions="variableSuggestions" :spellcheck="false" /></label>
      <div class="settings-grid compact-grid">
        <label><span>{{ t.proxyMode }}</span><select v-model="wsDraft.proxy.mode"><option value="inherit">{{ t.proxyInherit }}</option><option value="none">{{ t.proxyNone }}</option><option value="custom">{{ t.proxyCustom }}</option></select></label>
        <label v-if="wsDraft.proxy.mode === 'custom'"><span>{{ t.proxyUrl }}</span><VariableSuggestInput v-model="wsDraft.proxy.url" :suggestions="variableSuggestions" placeholder="socks5://127.0.0.1:10808" /></label>
      </div>
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
        <button class="send-btn" :disabled="realtimeBusy" @click="emit('runSSE')">
          <Loader2 v-if="realtimeBusy" class="spin" :size="15" />
          <Activity v-else :size="15" />
          {{ t.listen }}
        </button>
      </div>
      <label class="stack-label inline"><span>URL</span><VariableSuggestInput v-model="sseDraft.url" input-class="field" :suggestions="variableSuggestions" placeholder="https://example.com/events" /></label>
      <div class="settings-grid compact-grid">
        <label><span>{{ t.proxyMode }}</span><select v-model="sseDraft.proxy.mode"><option value="inherit">{{ t.proxyInherit }}</option><option value="none">{{ t.proxyNone }}</option><option value="custom">{{ t.proxyCustom }}</option></select></label>
        <label v-if="sseDraft.proxy.mode === 'custom'"><span>{{ t.proxyUrl }}</span><VariableSuggestInput v-model="sseDraft.proxy.url" :suggestions="variableSuggestions" placeholder="http://127.0.0.1:7890" /></label>
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
