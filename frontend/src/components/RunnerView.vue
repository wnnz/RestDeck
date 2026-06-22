<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircle2, Loader2, Play, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'

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
  runnerBusy: boolean
}>()

const emit = defineEmits<{
  selectCollection: [id: string]
  selectRequest: [id: string]
  setScope: [scope: 'collection' | 'request']
  setIterations: [iterations: number]
  runCollection: []
  runRequest: []
}>()

const collectionRequests = computed(() => props.activeCollection?.requests ?? [])
const selectedRequest = computed(() => collectionRequests.value.find((request) => request.id === props.activeRequestId) ?? collectionRequests.value[0] ?? null)
const canRunCollection = computed(() => !!props.activeCollection && (props.activeCollection.requests?.length ?? 0) > 0)
const canRunRequest = computed(() => !!selectedRequest.value)
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.runnerTitle }}</h2><p>{{ t.runnerHelp }}</p></div>
    <button class="send-btn" :disabled="runnerBusy || (runnerScope === 'collection' ? !canRunCollection : !canRunRequest)" @click="runnerScope === 'collection' ? emit('runCollection') : emit('runRequest')">
      <Loader2 v-if="runnerBusy" class="spin" :size="15" />
      <Play v-else :size="15" />
      {{ t.run }}
    </button>
  </div>

  <div class="runner-layout">
    <section class="runner-config">
      <div class="runner-card-title">
        <strong>{{ t.runnerTarget }}</strong>
        <span>{{ activeEnvironment?.name ?? t.activeEnvironment }}</span>
      </div>

      <div class="runner-scope-toggle">
        <button :class="{ active: runnerScope === 'collection' }" type="button" @click="emit('setScope', 'collection')">{{ t.runCollection }}</button>
        <button :class="{ active: runnerScope === 'request' }" type="button" @click="emit('setScope', 'request')">{{ t.runRequest }}</button>
      </div>

      <label class="runner-field">
        <span>{{ t.collections }}</span>
        <select :value="activeCollectionId" @change="emit('selectCollection', ($event.target as HTMLSelectElement).value)">
          <option v-for="collection in collections" :key="collection.id" :value="collection.id">{{ collection.name }}</option>
        </select>
      </label>

      <label v-if="runnerScope === 'request'" class="runner-field">
        <span>{{ t.request }}</span>
        <select :value="selectedRequest?.id ?? ''" @change="emit('selectRequest', ($event.target as HTMLSelectElement).value)">
          <option v-for="request in collectionRequests" :key="request.id" :value="request.id">{{ request.method }} {{ request.name || request.url }}</option>
        </select>
      </label>

      <label v-else class="runner-field">
        <span>{{ t.iterations }}</span>
        <input :value="runnerIterations" min="1" step="1" type="number" @input="emit('setIterations', Number(($event.target as HTMLInputElement).value))" />
      </label>

      <div class="runner-target-summary">
        <div><span>{{ t.activeEnvironment }}</span><strong>{{ activeEnvironment?.name ?? '-' }}</strong></div>
        <div><span>{{ t.requestCount }}</span><strong>{{ activeCollection?.requests?.length ?? 0 }}</strong></div>
        <div v-if="runnerScope === 'request'"><span>{{ t.requestUrl }}</span><strong>{{ selectedRequest?.url ?? '-' }}</strong></div>
      </div>

      <button class="send-btn runner-run-wide" :disabled="runnerBusy || (runnerScope === 'collection' ? !canRunCollection : !canRunRequest)" @click="runnerScope === 'collection' ? emit('runCollection') : emit('runRequest')">
        <Loader2 v-if="runnerBusy" class="spin" :size="15" />
        <Play v-else :size="15" />
        {{ runnerScope === 'collection' ? t.runCollection : t.runRequest }}
      </button>
    </section>

    <section class="runner-results">
      <div class="runner-summary">
        <div><strong>{{ runnerResult?.passed ?? 0 }}</strong><span>{{ t.passed }}</span></div>
        <div><strong>{{ runnerResult?.failed ?? 0 }}</strong><span>{{ t.failed }}</span></div>
        <div><strong>{{ runnerResult ? `${runnerResult.durationMs} ms` : '-' }}</strong><span>{{ t.duration }}</span></div>
      </div>
      <div class="response-panel standalone runner-result-panel">
        <div v-for="item in runnerResult?.items ?? []" :key="item.name + item.message" class="test-row">
          <CheckCircle2 v-if="item.passed" :size="15" class="text-emerald-600" />
          <XCircle v-else :size="15" class="text-red-600" />
          <span>{{ item.name }}</span>
          <code v-if="item.message">{{ item.message }}</code>
        </div>
        <div v-if="!runnerResult" class="empty-panel">{{ t.runnerEmpty }}</div>
      </div>
    </section>
  </div>
</template>
