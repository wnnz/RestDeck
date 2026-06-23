<script setup lang="ts">
import Checkbox from 'primevue/checkbox'
import { computed } from 'vue'
import { cn } from '../../utils/classNames'

const props = defineProps<{
  modelValue: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  click: [event: MouseEvent]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (next) => emit('update:modelValue', Boolean(next))
})
</script>

<template>
  <Checkbox
    v-model="value"
    binary
    :disabled="disabled"
    :pt="{
      root: {
        class: cn('volt-checkbox relative inline-flex h-4 w-4 shrink-0 items-center justify-center align-middle'),
        onClick: (event: MouseEvent) => emit('click', event)
      },
      box: {
        class: cn(
          'flex h-4 w-4 items-center justify-center rounded border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--on-accent)] transition-colors',
          modelValue && 'border-[var(--accent)] bg-[var(--accent)]',
          disabled && 'cursor-not-allowed opacity-60'
        )
      },
      input: {
        class: 'absolute inset-0 h-4 w-4 cursor-pointer opacity-0',
        onBlur: (event: FocusEvent) => emit('blur', event),
        onFocus: (event: FocusEvent) => emit('focus', event)
      },
      icon: { class: 'h-[11px] w-[11px]' }
    }"
    unstyled
  />
</template>
