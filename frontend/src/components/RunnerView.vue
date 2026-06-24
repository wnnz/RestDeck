<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircle2, Clipboard, Clock3, Download, Loader2, OctagonMinus, Play, Square, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { RunnerFailurePolicy, RunnerQueueItem } from '../types'
import VoltSelect from './volt/VoltSelect.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltInputText from './volt/VoltInputText.vue'
import VoltSelectButton from './volt/VoltSelectButton.vue'

const props = defineProps<{
  t: Translation
  collections: domain.Collection[]
  activeCollectionId: string
  activeRequestId: string
  runnerScope: 'collection' | 'request'
  runnerIterations: number
  activeEnvironment: domain.Environment | null
  activeCollection: domain.Collection | null
  runnerResult: domain.RunnerResult | null
  runnerQueue: RunnerQueueItem[]
  runnerBusy: boolean
  runnerFailurePolicy: RunnerFailurePolicy
}>()

const emit = defineEmits<{
  selectCollection: [id: string]
  selectRequest: [id: string]
  setScope: [scope: 'collection' | 'request']
  setIterations: [iterations: number]
  setFailurePolicy: [policy: RunnerFailurePolicy]
  runCollection: []
  runRequest: []
  stopRun: []
  copyReport: []
  exportReport: []
}>()

const collectionRequests = computed(() => props.activeCollection?.requests ?? [])
const selectedRequest = computed(() => collectionRequests.value.find((request) => request.id === props.activeRequestId) ?? collectionRequests.value[0] ?? null)
const collectionOptions = computed(() => props.collections.map((collection) => ({ value: collection.id, label: collection.name })))
const requestOptions = computed(() => collectionRequests.value.map((request) => ({ value: request.id, label: `${request.method} ${request.name || request.url}` })))
const canRunCollection = computed(() => !!props.activeCollection && (props.activeCollection.requests?.length ?? 0) > 0)
const canRunRequest = computed(() => !!selectedRequest.value)
const plannedCount = computed(() => props.runnerScope === 'collection' ? (props.activeCollection?.requests?.length ?? 0) * props.runnerIterations : (selectedRequest.value ? 1 : 0))
const scopeOptions = computed(() => [
  { value: 'collection', label: props.t.runCollection },
  { value: 'request', label: props.t.runRequest }
])
const failurePolicyOptions = computed(() => [
  { value: 'continue', label: props.t.continueOnFailure },
  { value: 'stopOnFailure', label: props.t.stopOnFailure }
])
const displayQueue = computed(() => props.runnerQueue.length ? props.runnerQueue : previewQueue.value)
const hasReport = computed(() => props.runnerQueue.length > 0 && !props.runnerBusy)
const previewQueue = computed<RunnerQueueItem[]>(() => {
  const requests = props.runnerScope === 'collection' ? collectionRequests.value : (selectedRequest.value ? [selectedRequest.value] : [])
  return Array.from({ length: props.runnerScope === 'collection' ? props.runnerIterations : 1 }).flatMap((_, iterationIndex) => requests.map((request) => ({
    id: `${request.id}-${iterationIndex + 1}`,
    requestId: request.id,
    iteration: iterationIndex + 1,
    method: request.method,
    name: request.name || request.url,
    url: request.url,
    status: 'waiting'
  })))
})

function statusLabel(status: RunnerQueueItem['status']) {
  if (status === 'running') return props.t.running
  if (status === 'passed') return props.t.passed
  if (status === 'failed') return props.t.failed
  if (status === 'skipped') return props.t.skipped
  return props.t.waiting
}
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.runnerTitle }}</h2><p>{{ t.runnerHelp }}</p></div>
  </div>

  <div class="runner-layout">
    <section class="runner-config">
      <div class="runner-card-title">
        <strong>{{ t.runnerTarget }}</strong>
        <span>{{ activeEnvironment?.name ?? t.activeEnvironment }}</span>
      </div>

      <VoltSelectButton class="runner-scope-toggle" :model-value="runnerScope" :options="scopeOptions" @update:model-value="emit('setScope', $event as 'collection' | 'request')" />

      <label class="runner-field">
        <span>{{ t.collections }}</span>
        <VoltSelect :model-value="activeCollectionId" :options="collectionOptions" @change="emit('selectCollection', String($event))" />
      </label>

      <label v-if="runnerScope === 'request'" class="runner-field">
        <span>{{ t.request }}</span>
        <VoltSelect :model-value="selectedRequest?.id ?? ''" :options="requestOptions" @change="emit('selectRequest', String($event))" />
      </label>

      <label v-else class="runner-field">
        <span>{{ t.iterations }}</span>
        <VoltInputText :model-value="runnerIterations" type="number" @update:model-value="emit('setIterations', Number($event))" />
      </label>

      <label class="runner-field">
        <span>{{ t.failurePolicy }}</span>
        <VoltSelect :model-value="runnerFailurePolicy" :options="failurePolicyOptions" @change="emit('setFailurePolicy', $event as RunnerFailurePolicy)" />
      </label>

      <div class="runner-target-summary">
        <div><span>{{ t.activeEnvironment }}</span><strong>{{ activeEnvironment?.name ?? '-' }}</strong></div>
        <div><span>{{ t.requestCount }}</span><strong>{{ plannedCount }}</strong></div>
        <div v-if="runnerScope === 'request'"><span>{{ t.requestUrl }}</span><strong>{{ selectedRequest?.url ?? '-' }}</strong></div>
      </div>

      <VoltButton v-if="runnerBusy" class="send-btn runner-run-wide runner-stop-button" @click="emit('stopRun')">
        <Square :size="15" />
        {{ t.stopRun }}
      </VoltButton>
      <VoltButton v-else class="send-btn runner-run-wide" :disabled="runnerScope === 'collection' ? !canRunCollection : !canRunRequest" @click="runnerScope === 'collection' ? emit('runCollection') : emit('runRequest')">
        <Loader2 v-if="runnerBusy" class="spin" :size="15" />
        <Play v-else :size="15" />
        {{ runnerScope === 'collection' ? t.startRunCollection : t.startRunRequest }}
      </VoltButton>
    </section>

    <section class="runner-results">
      <div class="runner-summary">
        <div><strong>{{ runnerResult?.passed ?? 0 }}</strong><span>{{ t.passed }}</span></div>
        <div><strong>{{ runnerResult?.failed ?? 0 }}</strong><span>{{ t.failed }}</span></div>
        <div><strong>{{ runnerResult ? `${runnerResult.durationMs} ms` : '-' }}</strong><span>{{ t.duration }}</span></div>
      </div>
      <div class="runner-actions">
        <VoltButton variant="secondary" :disabled="!hasReport" @click="emit('copyReport')"><Clipboard :size="14" /> {{ t.copyReport }}</VoltButton>
        <VoltButton variant="secondary" :disabled="!hasReport" @click="emit('exportReport')"><Download :size="14" /> {{ t.exportReport }}</VoltButton>
      </div>
      <div class="response-panel standalone runner-result-panel">
        <div v-if="displayQueue.length" class="runner-queue-row runner-queue-header">
          <span></span>
          <span>{{ t.method }}</span>
          <span>{{ t.request }}</span>
          <span>{{ t.status }}</span>
          <span>{{ t.result }}</span>
          <span>{{ t.duration }}</span>
          <span>{{ t.description }}</span>
        </div>
        <div v-for="item in displayQueue" :key="item.id" class="runner-queue-row">
          <CheckCircle2 v-if="item.status === 'passed'" :size="15" class="text-emerald-600" />
          <XCircle v-else-if="item.status === 'failed'" :size="15" class="text-red-600" />
          <Loader2 v-else-if="item.status === 'running'" class="spin" :size="15" />
          <OctagonMinus v-else-if="item.status === 'skipped'" :size="15" class="muted" />
          <Clock3 v-else :size="15" class="muted" />
          <span :class="['method', item.method.toLowerCase()]">{{ item.method }}</span>
          <div class="runner-queue-main">
            <strong>{{ item.name }}</strong>
            <small>{{ item.url }}</small>
          </div>
          <span :class="['runner-status', item.status]">{{ statusLabel(item.status) }}</span>
          <code>{{ item.statusCode ? `${item.statusCode}` : '-' }}</code>
          <span>{{ item.durationMs != null ? `${item.durationMs} ms` : '-' }}</span>
          <code>{{ item.message || '-' }}</code>
        </div>
        <div v-if="!displayQueue.length" class="empty-panel">{{ t.runnerEmpty }}</div>
      </div>
    </section>
  </div>
</template>
