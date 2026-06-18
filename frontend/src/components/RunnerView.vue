<script setup lang="ts">
import { CheckCircle2, Loader2, Play, XCircle } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'

defineProps<{
  t: Translation
  activeCollection: domain.Collection | null
  runnerResult: domain.RunnerResult | null
  runnerBusy: boolean
}>()

const emit = defineEmits<{
  runCollection: []
}>()
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.runnerTitle }}</h2><p>{{ activeCollection?.name ?? t.collections }}</p></div>
    <button class="send-btn" :disabled="runnerBusy || !activeCollection" @click="emit('runCollection')">
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
