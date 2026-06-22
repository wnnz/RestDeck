<script setup lang="ts">
import SelectButton from 'primevue/selectbutton'
import { computed } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

export type VoltSelectButtonOption = {
  value: string | number
  label: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: string | number
  options: VoltSelectButtonOption[]
  allowEmpty?: boolean
  class?: ClassValue
}>(), {
  allowEmpty: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  change: [value: string | number]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (next) => {
    if (next == null) return
    emit('update:modelValue', next)
    emit('change', next)
  }
})

const pt = computed(() => ({
  root: {
    class: cn('volt-select-button grid gap-1.5', props.class)
  },
  pcToggleButton: {
    root: ({ context }: { context: { active: boolean; disabled: boolean } }) => ({
      class: cn(
        'inline-flex h-[32px] min-w-0 items-center justify-center rounded border border-[var(--input-border)] bg-[var(--panel)] px-3 text-[13px] text-[var(--text)] outline-none transition-colors',
        'hover:bg-[var(--hover-bg)] focus-visible:ring-2 focus-visible:ring-[color-mix(in_srgb,var(--accent)_20%,transparent)]',
        context.active && 'border-[var(--accent-soft)] bg-[var(--accent-bg)] text-[var(--accent-text)] font-bold',
        context.disabled && 'cursor-not-allowed opacity-60'
      )
    }),
    label: { class: 'truncate' }
  }
}))
</script>

<template>
  <SelectButton
    v-model="value"
    :options="options"
    option-label="label"
    option-value="value"
    option-disabled="disabled"
    :allow-empty="allowEmpty"
    :pt="pt"
    unstyled
  />
</template>
