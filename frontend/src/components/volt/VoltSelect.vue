<script setup lang="ts">
import Select from 'primevue/select'
import { computed } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

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
  class?: ClassValue
  overlayClass?: ClassValue
}>(), {
  placeholder: ''
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  change: [value: string | number]
  click: [event: MouseEvent]
  dblclick: [event: MouseEvent]
  keydown: [event: KeyboardEvent]
  blur: [event: FocusEvent]
  focus: [event: FocusEvent]
}>()

const selectValue = computed({
  get: () => props.modelValue,
  set: (value) => {
    const next = value ?? ''
    emit('update:modelValue', next)
    emit('change', next)
  }
})

const pt = computed(() => ({
  root: {
    onClick: (event: MouseEvent) => emit('click', event),
    onDblclick: (event: MouseEvent) => emit('dblclick', event),
    onKeydown: (event: KeyboardEvent) => emit('keydown', event),
    onBlur: (event: FocusEvent) => emit('blur', event),
    onFocus: (event: FocusEvent) => emit('focus', event),
    class: cn(
      'volt-select inline-flex h-[34px] min-w-0 items-center rounded-md border border-[var(--input-border)] bg-[var(--input-bg)] text-[13px] text-[var(--text)] shadow-none outline-none transition-colors',
      'hover:border-[var(--border-strong)] focus-within:border-[var(--accent)] focus-within:ring-2 focus-within:ring-[color-mix(in_srgb,var(--accent)_18%,transparent)]',
      props.disabled && 'cursor-not-allowed opacity-60',
      props.class
    )
  },
  label: {
    class: 'min-w-0 flex-1 truncate px-3 py-0 leading-[32px] outline-none'
  },
  dropdown: {
    class: 'flex h-full w-8 shrink-0 items-center justify-center text-[var(--muted)] outline-none'
  },
  dropdownIcon: {
    class: 'h-[14px] w-[14px]'
  },
  overlay: {
    class: cn(
      'volt-select-overlay z-[1000] mt-1 overflow-hidden rounded-md border border-[var(--border)] bg-[var(--panel)] py-1 shadow-[0_16px_38px_var(--shadow)]',
      props.overlayClass
    )
  },
  listContainer: {
    class: 'max-h-[260px] overflow-auto'
  },
  list: {
    class: 'm-0 list-none p-1 outline-none'
  },
  option: ({ context }: { context: { selected: boolean; focused: boolean; disabled: boolean } }) => ({
    class: cn(
      'flex min-h-[30px] cursor-pointer items-center gap-2 rounded px-2.5 py-1.5 text-[13px] text-[var(--text)] outline-none',
      'hover:bg-[var(--hover-bg)]',
      context.focused && 'bg-[var(--hover-bg)]',
      context.selected && 'bg-[var(--accent-bg)] text-[var(--accent-text)]',
      context.disabled && 'cursor-not-allowed opacity-50'
    )
  }),
  optionLabel: {
    class: 'min-w-0 flex-1 truncate'
  },
  emptyMessage: {
    class: 'px-3 py-2 text-[13px] text-[var(--muted)]'
  },
  hiddenFirstFocusableEl: {
    class: 'outline-none'
  },
  hiddenFilterResult: {
    class: 'outline-none'
  },
  hiddenEmptyMessage: {
    class: 'outline-none'
  },
  hiddenSelectedMessage: {
    class: 'outline-none'
  },
  hiddenLastFocusableEl: {
    class: 'outline-none'
  }
}))
</script>

<template>
  <Select
    v-model="selectValue"
    :options="options"
    option-label="label"
    option-value="value"
    option-disabled="disabled"
    :placeholder="placeholder"
    :disabled="disabled"
    append-to="body"
    :pt="pt"
    unstyled
  >
    <template #option="{ option }">
      <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
      <small v-if="option.detail" class="truncate text-[12px] text-[var(--muted)]">{{ option.detail }}</small>
    </template>
  </Select>
</template>
