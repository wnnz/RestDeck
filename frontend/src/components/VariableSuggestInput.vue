<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { VariableSuggestion } from '../types'
import VoltInputText from './volt/VoltInputText.vue'
import VoltTextarea from './volt/VoltTextarea.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltPopover from './volt/VoltPopover.vue'

const props = withDefaults(defineProps<{
  modelValue: string | number | undefined
  suggestions: VariableSuggestion[]
  as?: 'input' | 'textarea'
  type?: string
  placeholder?: string
  wrapperClass?: string
  inputClass?: string
  spellcheck?: boolean
}>(), {
  as: 'input',
  type: 'text',
  placeholder: '',
  spellcheck: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  scroll: [event: Event]
}>()

const inputRef = ref<InstanceType<typeof VoltInputText> | InstanceType<typeof VoltTextarea> | null>(null)
const suggestPopover = ref<InstanceType<typeof VoltPopover> | null>(null)
const open = ref(false)
const query = ref('')
const selectedIndex = ref(0)

const modelText = computed(() => props.modelValue == null ? '' : String(props.modelValue))
const filteredSuggestions = computed(() => {
  const keyword = query.value.toLowerCase()
  const source = props.suggestions ?? []
  if (!keyword) return source.slice(0, 9)
  return source
    .filter((item) => item.name.toLowerCase().includes(keyword) || item.detail.toLowerCase().includes(keyword))
    .slice(0, 9)
})

watch(open, (next) => {
  const target = inputElement()
  if (!next || !target || !filteredSuggestions.value.length) {
    suggestPopover.value?.hide()
    return
  }
  const event = new Event('variablesuggest')
  void nextTick(() => suggestPopover.value?.show(event, target))
})

watch(filteredSuggestions, (items) => {
  if (!items.length) {
    suggestPopover.value?.hide()
    return
  }
  if (open.value) {
    const target = inputElement()
    if (target) {
      const event = new Event('variablesuggest')
      void nextTick(() => suggestPopover.value?.show(event, target))
    }
  }
})

function updateValue(value: string) {
  emit('update:modelValue', props.type === 'number' ? Number(value) : value)
}

function inputElement() {
  return inputRef.value?.input ?? null
}

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement | HTMLTextAreaElement
  updateValue(target.value)
  updateSuggestionState(target)
}

function updateSuggestionState(target = inputElement()) {
  if (!target) return
  const caret = target.selectionStart ?? target.value.length
  const beforeCaret = target.value.slice(0, caret)
  const trigger = beforeCaret.lastIndexOf('{{')
  if (trigger < 0 || beforeCaret.slice(trigger + 2).includes('}}')) {
    open.value = false
    return
  }
  query.value = beforeCaret.slice(trigger + 2).trimStart()
  selectedIndex.value = 0
  open.value = true
}

function insertSuggestion(item: VariableSuggestion) {
  const target = inputElement()
  if (!target) return
  const value = target.value
  const caret = target.selectionStart ?? value.length
  const beforeCaret = value.slice(0, caret)
  const trigger = beforeCaret.lastIndexOf('{{')
  if (trigger < 0) return
  const next = `${value.slice(0, trigger)}{{${item.name}}}${value.slice(caret)}`
  updateValue(next)
  open.value = false
  suggestPopover.value?.hide()
  void nextTick(() => {
    const pos = trigger + item.name.length + 4
    target.focus()
    target.setSelectionRange(pos, pos)
  })
}

function handleKeydown(event: KeyboardEvent) {
  if (!open.value) {
    if (event.key === '{') {
      void nextTick(() => updateSuggestionState())
    }
    return
  }
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filteredSuggestions.value.length - 1)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (event.key === 'Enter' || event.key === 'Tab') {
    const item = filteredSuggestions.value[selectedIndex.value]
    if (item) {
      event.preventDefault()
      insertSuggestion(item)
    }
  } else if (event.key === 'Escape') {
    open.value = false
    suggestPopover.value?.hide()
  }
}
</script>

<template>
  <span :class="['variable-input-wrap', wrapperClass]">
    <VoltTextarea
      v-if="as === 'textarea'"
      ref="inputRef"
      :model-value="modelText"
      :input-class="inputClass"
      :placeholder="placeholder"
      :spellcheck="spellcheck"
      @blur="open = false"
      @click="updateSuggestionState()"
      @input="handleInput"
      @update:model-value="updateValue(String($event))"
      @keydown="handleKeydown"
      @scroll="emit('scroll', $event)"
    />
    <VoltInputText
      v-else
      ref="inputRef"
      :input-class="inputClass"
      :type="type"
      :model-value="modelText"
      :placeholder="placeholder"
      :spellcheck="spellcheck"
      @blur="open = false"
      @click="updateSuggestionState()"
      @input="handleInput"
      @update:model-value="updateValue(String($event))"
      @keydown="handleKeydown"
    />
    <VoltPopover ref="suggestPopover" class="variable-suggest-popover" content-class="variable-suggest" @mousedown.prevent>
      <VoltButton
        v-for="(item, index) in filteredSuggestions"
        :key="item.name"
        :class="{ active: index === selectedIndex }"
        variant="ghost"
        @mousedown.prevent="insertSuggestion(item)"
      >
        <strong>{{ item.name }}</strong>
        <span>{{ item.detail }}</span>
      </VoltButton>
    </VoltPopover>
  </span>
</template>
