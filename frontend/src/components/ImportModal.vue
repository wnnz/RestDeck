<script setup lang="ts">
import { Download, FolderOpen, Import, Link, Loader2, X } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { ActiveModal } from '../types'
import VoltButton from './volt/VoltButton.vue'
import VoltDialog from './volt/VoltDialog.vue'
import VoltInputText from './volt/VoltInputText.vue'
import VoltSelect from './volt/VoltSelect.vue'
import VoltTextarea from './volt/VoltTextarea.vue'

defineProps<{
  activeModal: ActiveModal
  title: string
  busy: boolean
  t: Translation
  postmanText: string
  openAPIText: string
  swaggerUrl: string
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
  'import-from-file': []
  'export-to-file': []
  'update:postmanText': [value: string]
  'update:openAPIText': [value: string]
  'update:swaggerUrl': [value: string]
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
  <VoltDialog :visible="!!activeModal" :class="['modal', { 'swagger-modal': activeModal === 'swagger' }]" @update:visible="updateVisible">
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
      :placeholder="t.postmanPlaceholder"
      @update:model-value="emit('update:postmanText', String($event))"
    />
    <template v-else-if="activeModal === 'openapi'">
      <div v-if="openAPIServers.length" class="modal-import-toolbar">
        <label class="modal-server-field">
          <span>{{ t.openAPIServer }}</span>
          <VoltSelect
            :model-value="selectedOpenAPIServer"
            :options="openAPIServers.map((server) => ({ value: server, label: server }))"
            @change="emit('update:selectedOpenAPIServer', String($event))"
          />
        </label>
      </div>
      <VoltTextarea
        :model-value="openAPIText"
        input-class="modal-textarea"
        :spellcheck="false"
        :aria-label="t.openAPIJSON"
        :placeholder="t.openAPIPlaceholder"
        @update:model-value="emit('update:openAPIText', String($event))"
      />
    </template>
    <div v-else-if="activeModal === 'swagger'" class="swagger-url-panel">
      <label class="swagger-url-field">
        <span>{{ t.swaggerUrl }}</span>
        <VoltInputText
          :model-value="swaggerUrl"
          input-class="swagger-url-input"
          :placeholder="t.swaggerUrlPlaceholder"
          @update:model-value="emit('update:swaggerUrl', String($event))"
          @keydown.enter="emit('submit')"
        />
      </label>
    </div>
    <VoltTextarea
      v-else-if="activeModal === 'har'"
      :model-value="harText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.harJSON"
      :placeholder="t.harPlaceholder"
      @update:model-value="emit('update:harText', String($event))"
    />
    <VoltTextarea
      v-else-if="activeModal === 'fetch'"
      :model-value="fetchText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.fetchSnippet"
      :placeholder="t.fetchPlaceholder"
      @update:model-value="emit('update:fetchText', String($event))"
    />
    <VoltTextarea
      v-else-if="activeModal === 'curl'"
      :model-value="curlText"
      input-class="modal-textarea"
      :spellcheck="false"
      :aria-label="t.curlSnippet"
      :placeholder="t.curlPlaceholder"
      @update:model-value="emit('update:curlText', String($event))"
    />
    <VoltTextarea v-else :model-value="exportText" input-class="modal-textarea" :spellcheck="false" readonly :aria-label="t.exportedCollection" />
    <div class="modal-actions">
      <VoltButton v-if="activeModal !== 'export' && activeModal !== 'swagger'" variant="secondary" :disabled="busy" @click="emit('import-from-file')">
        <FolderOpen :size="15" />
        {{ t.importFromFile }}
      </VoltButton>
      <span v-else />
      <VoltButton v-if="activeModal !== 'export'" class="send-btn" :disabled="busy" @click="emit('submit')">
        <Loader2 v-if="busy" class="spin" :size="15" />
        <Link v-else-if="activeModal === 'swagger'" :size="15" />
        <Import v-else :size="15" />
        {{ title }}
      </VoltButton>
      <VoltButton v-else class="send-btn" :disabled="!exportText" @click="emit('export-to-file')">
        <Download :size="15" />
        {{ t.exportToFile }}
      </VoltButton>
    </div>
  </VoltDialog>
</template>
