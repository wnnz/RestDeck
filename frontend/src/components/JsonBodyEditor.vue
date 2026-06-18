<script setup lang="ts">
import { computed } from 'vue'
import { tokenizeJSON } from '../utils/jsonHighlight'

const props = defineProps<{
  modelValue: string
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
    <textarea
      class="json-body-input"
      :value="modelValue"
      spellcheck="false"
      placeholder='{"hello": "world"}'
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      @scroll="syncScroll"
    ></textarea>
    <pre class="json-body-highlight" aria-hidden="true"><span v-for="(token, index) in tokens" :key="index" :class="`json-${token.type}`">{{ token.text }}</span></pre>
  </div>
</template>
