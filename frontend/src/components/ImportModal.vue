<script setup lang="ts">
import { Import, Loader2, X } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { ActiveModal } from '../types'
import VoltButton from './volt/VoltButton.vue'
import VoltDialog from './volt/VoltDialog.vue'
import VoltTextarea from './volt/VoltTextarea.vue'

defineProps<{
  activeModal: ActiveModal
  title: string
  busy: boolean
  t: Translation
  postmanText: string
  fetchText: string
  curlText: string
  exportText: string
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'update:postmanText': [value: string]
  'update:fetchText': [value: string]
  'update:curlText': [value: string]
}>()

function updateVisible(value: boolean) {
  if (!value) emit('close')
}
</script>

<template>
  <VoltDialog :visible="!!activeModal" class="modal" @update:visible="updateVisible">
    <div class="modal-title">
      <strong>{{ title }}</strong>
      <VoltButton class="modal-close" size="icon" variant="ghost" title="Close" @click="emit('close')">
        <X :size="14" :stroke-width="1.7" />
      </VoltButton>
    </div>
    <VoltTextarea
      v-if="activeModal === 'postman'"
      :model-value="postmanText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.postmanJSON"
      @update:model-value="emit('update:postmanText', String($event))"
    />
    <VoltTextarea
      v-else-if="activeModal === 'fetch'"
      :model-value="fetchText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.fetchSnippet"
      @update:model-value="emit('update:fetchText', String($event))"
    />
    <VoltTextarea
      v-else-if="activeModal === 'curl'"
      :model-value="curlText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.curlSnippet"
      @update:model-value="emit('update:curlText', String($event))"
    />
    <VoltTextarea v-else :model-value="exportText" input-class="modal-textarea" :spellcheck="false" readonly :aria-label="t.exportedCollection" />
    <VoltButton v-if="activeModal !== 'export'" class="send-btn" :disabled="busy" @click="emit('submit')">
      <Loader2 v-if="busy" class="spin" :size="15" />
      <Import v-else :size="15" />
      {{ title }}
    </VoltButton>
  </VoltDialog>
</template>
