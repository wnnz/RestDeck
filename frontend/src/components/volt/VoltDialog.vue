<script setup lang="ts">
import Dialog from 'primevue/dialog'
import { computed } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

const props = withDefaults(defineProps<{
  visible: boolean
  modal?: boolean
  dismissableMask?: boolean
  blockScroll?: boolean
  class?: ClassValue
  contentClass?: ClassValue
  maskClass?: ClassValue
}>(), {
  modal: true,
  dismissableMask: true,
  blockScroll: true
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const rootClass = computed(() => cn(
  'volt-dialog flex max-h-[calc(100vh-48px)] min-h-0 flex-col overflow-hidden rounded-md border border-[var(--border)] bg-[var(--panel)] text-[var(--text)] shadow-[0_24px_70px_var(--shadow-strong)]',
  props.class
))

const maskClass = computed(() => cn(
  'volt-dialog-mask fixed inset-0 z-[1100] flex items-center justify-center bg-[var(--overlay)] p-6',
  props.maskClass
))

const contentClass = computed(() => cn(
  'flex min-h-0 flex-1 flex-col',
  props.contentClass
))
</script>

<template>
  <Dialog
    :visible="visible"
    :modal="modal"
    :dismissable-mask="dismissableMask"
    :block-scroll="blockScroll"
    append-to="body"
    :show-header="false"
    :pt="{
      mask: { class: maskClass },
      root: { class: rootClass },
      content: { class: contentClass }
    }"
    unstyled
    @update:visible="emit('update:visible', $event)"
  >
    <slot />
  </Dialog>
</template>
