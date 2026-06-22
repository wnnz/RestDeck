<script setup lang="ts">
import { computed } from 'vue'
import { tokenizeJSON } from '../utils/jsonHighlight'
import type { VariableSuggestion } from '../types'
import VariableSuggestInput from './VariableSuggestInput.vue'

const props = defineProps<{
  modelValue: string
  suggestions: VariableSuggestion[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const tokens = computed(() => tokenizeJSON(props.modelValue || ''))

function syncScroll(event: Event) {
  const target = event.target as HTMLTextAreaElement
  const highlight = target.nextElementSibling as HTMLElement | null
  if (!highlight) return
  highlight.scrollTop = target.scrollTop
  highlight.scrollLeft = target.scrollLeft
}
</script>

<template>
  <div class="json-body-editor">
    <VariableSuggestInput
      input-class="json-body-input"
      as="textarea"
      :model-value="modelValue"
      :suggestions="suggestions"
      :spellcheck="false"
      placeholder='{"hello": "world"}'
      @update:model-value="emit('update:modelValue', String($event))"
      @scroll="syncScroll"
    />
    <pre class="json-body-highlight" aria-hidden="true"><span v-for="(token, index) in tokens" :key="index" :class="`json-${token.type}`">{{ token.text }}</span></pre>
  </div>
</template>
