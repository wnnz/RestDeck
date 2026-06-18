<script setup lang="ts">
import { Import, Loader2, X } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { ActiveModal } from '../types'

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
</script>

<template>
  <Teleport to="body">
    <div v-if="activeModal" class="modal-backdrop" @click.self="emit('close')">
      <section class="modal">
        <div class="modal-title">
          <strong>{{ title }}</strong>
          <button class="modal-close" type="button" title="Close" @click="emit('close')">
            <X :size="14" :stroke-width="1.7" />
          </button>
        </div>
        <textarea
          v-if="activeModal === 'postman'"
          :value="postmanText"
          spellcheck="false"
          :aria-label="t.postmanJSON"
          @input="emit('update:postmanText', ($event.target as HTMLTextAreaElement).value)"
        ></textarea>
        <textarea
          v-else-if="activeModal === 'fetch'"
          :value="fetchText"
          spellcheck="false"
          :aria-label="t.fetchSnippet"
          @input="emit('update:fetchText', ($event.target as HTMLTextAreaElement).value)"
        ></textarea>
        <textarea
          v-else-if="activeModal === 'curl'"
          :value="curlText"
          spellcheck="false"
          :aria-label="t.curlSnippet"
          @input="emit('update:curlText', ($event.target as HTMLTextAreaElement).value)"
        ></textarea>
        <textarea v-else :value="exportText" spellcheck="false" readonly :aria-label="t.exportedCollection"></textarea>
        <button v-if="activeModal !== 'export'" class="send-btn" :disabled="busy" @click="emit('submit')">
          <Loader2 v-if="busy" class="spin" :size="15" />
          <Import v-else :size="15" />
          {{ title }}
        </button>
      </section>
    </div>
  </Teleport>
</template>
