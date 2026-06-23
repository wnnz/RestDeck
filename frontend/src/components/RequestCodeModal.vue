<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Clipboard, X } from 'lucide-vue-next'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import { codeFormatOptions, generateRequestCode, type CodeFormat } from '../utils/codeGenerator'
import VoltButton from './volt/VoltButton.vue'
import VoltDialog from './volt/VoltDialog.vue'
import VoltSelect from './volt/VoltSelect.vue'
import VoltTextarea from './volt/VoltTextarea.vue'

const props = defineProps<{
  visible: boolean
  request: domain.Request | null
  t: Translation
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  copied: []
}>()

const selectedFormat = ref<CodeFormat>('curl-linux')

const generatedCode = computed(() => {
  if (!props.request) return ''
  return generateRequestCode(props.request, selectedFormat.value)
})

watch(() => props.visible, (visible) => {
  if (visible) selectedFormat.value = 'curl-linux'
})

function updateVisible(value: boolean) {
  emit('update:visible', value)
}

async function copyCode() {
  if (!generatedCode.value) return
  await ClipboardSetText(generatedCode.value)
  emit('copied')
}
</script>

<template>
  <VoltDialog :visible="visible" class="modal code-modal" @update:visible="updateVisible">
    <div class="modal-title">
      <strong>{{ t.generateCode }}</strong>
      <VoltButton class="modal-close" size="icon" variant="ghost" title="Close" @click="emit('update:visible', false)">
        <X :size="14" :stroke-width="1.7" />
      </VoltButton>
    </div>

    <div class="code-modal-toolbar">
      <label class="settings-field code-format-field">
        <span>{{ t.codeFormat }}</span>
        <VoltSelect v-model="selectedFormat" :options="codeFormatOptions" />
      </label>
      <VoltButton class="send-btn" @click="copyCode">
        <Clipboard :size="15" />
        {{ t.copyCode }}
      </VoltButton>
    </div>

    <VoltTextarea
      :model-value="generatedCode"
      input-class="modal-textarea code-preview"
      :spellcheck="false"
      readonly
      :aria-label="t.generatedCode"
    />
  </VoltDialog>
</template>
