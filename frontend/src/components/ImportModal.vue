<script setup lang="ts">
import { Import, Loader2, X } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { ActiveModal } from '../types'
import VoltButton from './volt/VoltButton.vue'
import VoltDialog from './volt/VoltDialog.vue'
import VoltSelect from './volt/VoltSelect.vue'
import VoltTextarea from './volt/VoltTextarea.vue'

defineProps<{
  activeModal: ActiveModal
  title: string
  busy: boolean
  t: Translation
  postmanText: string
  openAPIText: string
  harText: string
  fetchText: string
  curlText: string
  exportText: string
  openAPIServers: string[]
  selectedOpenAPIServer: string
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'update:postmanText': [value: string]
  'update:openAPIText': [value: string]
  'update:harText': [value: string]
  'update:selectedOpenAPIServer': [value: string]
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
    <template v-else-if="activeModal === 'openapi'">
      <label v-if="openAPIServers.length" class="modal-field">
        <span>{{ t.openAPIServer }}</span>
        <VoltSelect
          :model-value="selectedOpenAPIServer"
          :options="openAPIServers.map((server) => ({ value: server, label: server }))"
          @change="emit('update:selectedOpenAPIServer', String($event))"
        />
      </label>
      <VoltTextarea
        :model-value="openAPIText"
        input-class="modal-textarea"
        :spellcheck="false"
        :aria-label="t.openAPIJSON"
        @update:model-value="emit('update:openAPIText', String($event))"
      />
    </template>
    <VoltTextarea
      v-else-if="activeModal === 'har'"
      :model-value="harText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.harJSON"
      @update:model-value="emit('update:harText', String($event))"
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
