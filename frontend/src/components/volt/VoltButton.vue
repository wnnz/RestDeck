<script setup lang="ts">
import PrimeButton from 'primevue/button'
import { computed } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

const props = withDefaults(defineProps<{
  label?: string
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'icon'
  class?: ClassValue
}>(), {
  type: 'button',
  variant: 'secondary',
  size: 'md'
})

const emit = defineEmits<{
  click: [event: MouseEvent]
  dblclick: [event: MouseEvent]
  mousedown: [event: MouseEvent]
  contextmenu: [event: MouseEvent]
}>()

const rootClass = computed(() => cn(
  'inline-flex shrink-0 items-center justify-center gap-1.5 rounded border text-[13px] font-medium outline-none transition-colors',
  'focus-visible:ring-2 focus-visible:ring-[color-mix(in_srgb,var(--accent)_20%,transparent)]',
  props.size === 'sm' && 'h-[28px] min-w-[30px] px-2',
  props.size === 'md' && 'h-[32px] min-w-[32px] px-3',
  props.size === 'icon' && 'h-[30px] w-[30px] p-0',
  props.variant === 'primary' && 'border-[var(--accent)] bg-[var(--accent)] text-[var(--on-accent)] hover:bg-[var(--accent-strong)]',
  props.variant === 'secondary' && 'border-[var(--input-border)] bg-[var(--panel)] text-[var(--text)] hover:bg-[var(--hover-bg)] hover:border-[var(--border-strong)]',
  props.variant === 'ghost' && 'border-transparent bg-transparent text-[var(--muted-strong)] hover:bg-[var(--hover-bg)] hover:text-[var(--text)]',
  props.variant === 'danger' && 'border-[#dc2626] bg-[#dc2626] text-[var(--on-danger)] hover:bg-[#b91c1c]',
  (props.disabled || props.loading) && 'cursor-not-allowed opacity-60',
  props.class
))
</script>

<template>
  <PrimeButton
    :type="type"
    :disabled="disabled"
    :loading="loading"
    :label="label"
    :pt="{ root: { class: rootClass }, label: { class: 'truncate' } }"
    unstyled
    @click="emit('click', $event)"
    @dblclick="emit('dblclick', $event)"
    @mousedown="emit('mousedown', $event)"
    @contextmenu="emit('contextmenu', $event)"
  >
    <template v-if="$slots.default" #default>
      <slot />
    </template>
    <template v-if="$slots.icon" #icon>
      <slot name="icon" />
    </template>
    <template v-if="$slots.loadingicon" #loadingicon>
      <slot name="loadingicon" />
    </template>
  </PrimeButton>
</template>
