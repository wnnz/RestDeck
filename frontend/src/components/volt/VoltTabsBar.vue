<script setup lang="ts">
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import Tabs from 'primevue/tabs'
import { computed } from 'vue'
import type { ClassValue } from 'vue'
import { cn } from '../../utils/classNames'

export type VoltTabItem = {
  key: string
  label: string
  count?: number
  disabled?: boolean
}

const props = defineProps<{
  modelValue: string
  items: VoltTabItem[]
  class?: ClassValue
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (next) => emit('update:modelValue', String(next))
})

function tabRootPt({ context }: { context: { active: boolean; disabled: boolean } }) {
  return {
    class: cn('tab', context.active && 'active', context.disabled && 'opacity-60 cursor-not-allowed')
  }
}
</script>

<template>
  <Tabs
    :value="value"
    :pt="{ root: { class: cn('volt-tabs-bar tabs', props.class) } }"
    unstyled
    @update:value="value = String($event)"
  >
    <TabList :pt="{ root: { class: 'contents' }, activeBar: { class: 'hidden' }, tabList: { class: 'contents' } }" unstyled>
      <Tab
        v-for="item in items"
        :key="item.key"
        :value="item.key"
        :disabled="item.disabled"
        :pt="{
          root: tabRootPt
        }"
        unstyled
      >
        {{ item.label }}
        <span v-if="item.count" class="count">{{ item.count }}</span>
      </Tab>
    </TabList>
  </Tabs>
</template>
