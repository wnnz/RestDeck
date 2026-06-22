<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ChevronDown } from 'lucide-vue-next'

export type SelectOption = {
  value: string | number
  label: string
  detail?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: string | number
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  buttonClass?: string
  menuClass?: string
}>(), {
  placeholder: '',
  buttonClass: '',
  menuClass: ''
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  change: [value: string | number]
}>()

const root = ref<HTMLElement | null>(null)
const button = ref<HTMLButtonElement | null>(null)
const menu = ref<HTMLElement | null>(null)
const open = ref(false)
const activeIndex = ref(0)
const menuStyle = ref<Record<string, string>>({})

const enabledOptions = computed(() => props.options.filter((option) => !option.disabled))
const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue))
const displayLabel = computed(() => selectedOption.value?.label || props.placeholder)

watch(open, (next) => {
  if (!next) return
  const index = props.options.findIndex((option) => option.value === props.modelValue && !option.disabled)
  activeIndex.value = index >= 0 ? index : Math.max(0, props.options.findIndex((option) => !option.disabled))
  void nextTick(updateMenuPosition)
})

function close() {
  open.value = false
}

function toggle() {
  if (props.disabled) return
  open.value = !open.value
}

function updateMenuPosition() {
  if (!open.value || !button.value) return
  const rect = button.value.getBoundingClientRect()
  const gap = 4
  const margin = 8
  const spaceBelow = window.innerHeight - rect.bottom - margin
  const spaceAbove = rect.top - margin
  const openUp = spaceBelow < 160 && spaceAbove > spaceBelow
  const available = Math.max(120, openUp ? spaceAbove - gap : spaceBelow - gap)
  const maxHeight = Math.min(260, available)
  menuStyle.value = {
    left: `${Math.max(margin, rect.left)}px`,
    width: `${rect.width}px`,
    maxHeight: `${maxHeight}px`,
    ...(openUp
      ? { bottom: `${Math.max(margin, window.innerHeight - rect.top + gap)}px`, top: 'auto' }
      : { top: `${rect.bottom + gap}px`, bottom: 'auto' })
  }
}

function selectOption(option: SelectOption) {
  if (option.disabled) return
  emit('update:modelValue', option.value)
  emit('change', option.value)
  close()
  void nextTick(() => button.value?.focus())
}

function move(delta: number) {
  if (!enabledOptions.value.length) return
  const current = props.options[activeIndex.value]
  const enabledIndex = Math.max(0, enabledOptions.value.findIndex((option) => option.value === current?.value))
  const nextEnabled = enabledOptions.value[(enabledIndex + delta + enabledOptions.value.length) % enabledOptions.value.length]
  activeIndex.value = props.options.findIndex((option) => option.value === nextEnabled.value)
}

function handleKeydown(event: KeyboardEvent) {
  if (props.disabled) return
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (!open.value) open.value = true
    else move(1)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (!open.value) open.value = true
    else move(-1)
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    if (!open.value) {
      open.value = true
      return
    }
    const option = props.options[activeIndex.value]
    if (option) selectOption(option)
  } else if (event.key === 'Escape') {
    event.preventDefault()
    close()
  } else if (event.key === 'Tab') {
    close()
  }
}

function handleDocumentPointer(event: MouseEvent) {
  const target = event.target
  if (target instanceof Node && root.value?.contains(target)) return
  close()
}

onMounted(() => {
  document.addEventListener('mousedown', handleDocumentPointer, true)
  window.addEventListener('resize', updateMenuPosition)
  window.addEventListener('scroll', updateMenuPosition, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleDocumentPointer, true)
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
})
</script>

<template>
  <div ref="root" :class="['custom-select', { open, disabled }]">
    <button
      ref="button"
      :class="['custom-select-button', buttonClass]"
      type="button"
      :disabled="disabled"
      aria-haspopup="listbox"
      :aria-expanded="open"
      @click="toggle"
      @keydown="handleKeydown"
    >
      <span :class="{ placeholder: !selectedOption }">{{ displayLabel }}</span>
      <ChevronDown class="custom-select-arrow" :size="14" />
    </button>
    <div v-if="open" ref="menu" :class="['custom-select-menu', menuClass]" :style="menuStyle" role="listbox">
      <button
        v-for="(option, index) in options"
        :key="option.value"
        type="button"
        role="option"
        :aria-selected="option.value === modelValue"
        :disabled="option.disabled"
        :class="['custom-select-option', { active: index === activeIndex, selected: option.value === modelValue }]"
        @mouseenter="activeIndex = index"
        @click="selectOption(option)"
      >
        <span>{{ option.label }}</span>
        <small v-if="option.detail">{{ option.detail }}</small>
      </button>
    </div>
  </div>
</template>
