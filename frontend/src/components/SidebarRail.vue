<script setup lang="ts">
import type { Component } from 'vue'
import type { NavKey } from '../types'
import VoltButton from './volt/VoltButton.vue'

defineProps<{
  items: Array<{ key: NavKey; label: string; icon: Component }>
  activeNav: NavKey
}>()

const emit = defineEmits<{
  'update:activeNav': [value: NavKey]
}>()
</script>

<template>
  <aside class="rail">
    <VoltButton
      v-for="item in items"
      :key="item.key"
      :class="['rail-button', { active: activeNav === item.key }]"
      :title="item.label"
      variant="ghost"
      @click="emit('update:activeNav', item.key)"
    >
      <component :is="item.icon" :size="17" />
      <span>{{ item.label }}</span>
    </VoltButton>
  </aside>
</template>
