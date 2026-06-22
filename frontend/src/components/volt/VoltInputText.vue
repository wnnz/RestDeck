<script setup lang="ts">
import InputText from 'primevue/inputtext'
import { computed, ref } from 'vue'
import { cn } from '../../utils/classNames'

const props = withDefaults(defineProps<{
  modelValue?: string | number
  type?: string
  placeholder?: string
  disabled?: boolean
  readonly?: boolean
  spellcheck?: boolean
  inputClass?: string
}>(), {
  type: 'text',
  placeholder: '',
  spellcheck: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  input: [event: Event]
  keydown: [event: KeyboardEvent]
  click: [event: MouseEvent]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
}>()

const value = computed({
  get: () => props.modelValue == null ? '' : String(props.modelValue),
  set: (next) => emit('update:modelValue', props.type === 'number' ? Number(next) : next)
})

const input = ref<HTMLInputElement | null>(null)

defineExpose({
  input
})

const rootClass = computed(() => cn(
  'volt-input h-[32px] min-w-0 rounded border border-[var(--input-border)] bg-[var(--input-bg)] px-2.5 text-[13px] text-[var(--text)] outline-none transition-colors',
  'placeholder:text-[var(--muted-soft)] hover:border-[var(--border-strong)] focus:border-[var(--accent-soft)] focus:ring-2 focus:ring-[color-mix(in_srgb,var(--accent)_18%,transparent)]',
  props.disabled && 'cursor-not-allowed opacity-60',
  props.inputClass
))
</script>

<template>
  <InputText
    ref="input"
    v-model="value"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :readonly="readonly"
    :spellcheck="spellcheck"
    :pt="{ root: { class: rootClass } }"
    unstyled
    @input="emit('input', $event)"
    @keydown="emit('keydown', $event)"
    @click="emit('click', $event)"
    @blur="emit('blur', $event)"
    @focus="emit('focus', $event)"
  />
</template>
