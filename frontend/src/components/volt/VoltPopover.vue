<script setup lang="ts">
import Popover from 'primevue/popover'
import { computed, ref } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

const props = withDefaults(defineProps<{
  class?: ClassValue
  contentClass?: ClassValue
  dismissable?: boolean
}>(), {
  dismissable: true
})

const emit = defineEmits<{
  show: []
  hide: []
  mousedown: [event: MouseEvent]
}>()

const popover = ref<InstanceType<typeof Popover> | null>(null)

const rootClass = computed(() => cn(
  'volt-popover z-[1000] overflow-hidden rounded-md border border-[var(--input-border)] bg-[var(--panel)] shadow-[0_16px_34px_var(--shadow)]',
  props.class
))

const contentClass = computed(() => cn(
  'min-w-0',
  props.contentClass
))

function toggle(event: Event, target?: unknown) {
  popover.value?.toggle(event, target)
}

function show(event: Event, target?: unknown) {
  popover.value?.show(event, target)
}

function hide() {
  popover.value?.hide()
}

defineExpose({
  toggle,
  show,
  hide
})
</script>

<template>
  <Popover
    ref="popover"
    append-to="body"
    :dismissable="dismissable"
    :pt="{
      root: { class: rootClass },
      content: { class: contentClass, onMousedown: (event: MouseEvent) => emit('mousedown', event) }
    }"
    unstyled
    @show="emit('show')"
    @hide="emit('hide')"
  >
    <slot />
  </Popover>
</template>
