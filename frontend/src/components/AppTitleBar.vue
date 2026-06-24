<script setup lang="ts">
import { Moon, Search, Square, Sun, X } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import appIcon from '../assets/app-icon.png'
import type { Translation } from '../i18n/messages'
import type { Theme } from '../types'
import VoltSelect from './volt/VoltSelect.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltInputText from './volt/VoltInputText.vue'

defineProps<{
  t: Translation
  search: string
  theme: Theme
  activeEnvironment: domain.Environment | null
  environments: domain.Environment[]
}>()

const emit = defineEmits<{
  'update:search': [value: string]
  toggleTheme: []
  selectEnvironment: [id: string]
  minimize: []
  toggleMaximize: []
  close: []
}>()
</script>

<template>
  <header class="topbar window-titlebar" @dblclick="emit('toggleMaximize')">
    <div class="window-title">
      <img class="window-title-icon" :src="appIcon" alt="" aria-hidden="true" />
      <span>RestDeck</span>
    </div>
    <div class="top-search" @dblclick.stop>
      <Search :size="14" />
      <VoltInputText input-class="top-search-input" :model-value="search" :placeholder="t.search" @update:model-value="emit('update:search', String($event))" />
    </div>
    <div class="top-spacer" />
    <VoltSelect
      class="env-select"
      :model-value="activeEnvironment?.id"
      :options="environments.map((env) => ({ value: env.id, label: env.name }))"
      @dblclick.stop
      @change="emit('selectEnvironment', String($event))"
    />
    <VoltButton class="top-theme-btn" size="sm" @dblclick.stop @click="emit('toggleTheme')">
      <Sun v-if="theme === 'dark'" :size="14" />
      <Moon v-else :size="14" />
      {{ theme === 'dark' ? t.light : t.dark }}
    </VoltButton>
    <div class="window-controls" @dblclick.stop>
      <VoltButton class="window-control" title="Minimize" variant="ghost" @click="emit('minimize')">
        <span class="minimize-mark"></span>
      </VoltButton>
      <VoltButton class="window-control" title="Maximize" variant="ghost" @click="emit('toggleMaximize')">
        <Square :size="11" :stroke-width="1.7" />
      </VoltButton>
      <VoltButton class="window-control close" title="Close" variant="ghost" @click="emit('close')">
        <X :size="14" :stroke-width="1.7" />
      </VoltButton>
    </div>
  </header>
</template>
