<script setup lang="ts">
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import { statusClass } from '../utils/format'

defineProps<{
  t: Translation
  history: domain.HistoryItem[]
}>()

const emit = defineEmits<{
  selectRequest: [request: domain.Request]
}>()
</script>

<template>
  <div class="section-header"><div><h2>{{ t.history }}</h2><p>SQLite</p></div></div>
  <div class="history-table">
    <div class="history-head"><span>METHOD</span><span>{{ t.response }}</span><span>{{ t.status }}</span><span>{{ t.duration }}</span></div>
    <button v-for="item in history" :key="item.id" class="history-line" @click="emit('selectRequest', item.request)">
      <span :class="['method', item.method.toLowerCase()]">{{ item.method }}</span>
      <span class="truncate">{{ item.url }}</span>
      <span :class="statusClass(item.statusCode)">{{ item.statusCode || '-' }}</span>
      <span>{{ item.durationMs }} ms</span>
    </button>
  </div>
</template>
