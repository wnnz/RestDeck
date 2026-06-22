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
        class: cn('inline-flex h-[14px] w-[14px] items-center justify-center'),
        onClick: (event: MouseEvent) => emit('click', event)
      },
      box: {
        class: cn(
          'flex h-[14px] w-[14px] items-center justify-center rounded border border-[var(--input-border)] bg-[var(--input-bg)] text-[var(--on-accent)] transition-colors',
          modelValue && 'border-[var(--accent)] bg-[var(--accent)]',
          disabled && 'cursor-not-allowed opacity-60'
        )
      },
      input: {
        class: 'absolute h-[14px] w-[14px] cursor-pointer opacity-0',
        onBlur: (event: FocusEvent) => emit('blur', event),
        onFocus: (event: FocusEvent) => emit('focus', event)
      },
      icon: { class: 'h-[10px] w-[10px]' }
    }"
    unstyled
  />
</template>
