<script setup lang="ts">
import { Moon, Search, Square, Sun, X } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { Theme } from '../types'

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
    <div class="window-title">RestDeck</div>
    <div class="top-search" @dblclick.stop>
      <Search :size="14" />
      <input :value="search" :placeholder="t.search" @input="emit('update:search', ($event.target as HTMLInputElement).value)" />
    </div>
    <div class="top-spacer" />
    <select class="env-select" :value="activeEnvironment?.id" @dblclick.stop @change="emit('selectEnvironment', ($event.target as HTMLSelectElement).value)">
      <option v-for="env in environments" :key="env.id" :value="env.id">{{ env.name }}</option>
    </select>
    <button type="button" class="top-theme-btn" @dblclick.stop @click="emit('toggleTheme')">
      <Sun v-if="theme === 'dark'" :size="14" />
      <Moon v-else :size="14" />
      {{ theme === 'dark' ? t.light : t.dark }}
    </button>
    <div class="window-controls" @dblclick.stop>
      <button type="button" class="window-control" title="Minimize" @click="emit('minimize')">
        <span class="minimize-mark"></span>
      </button>
      <button type="button" class="window-control" title="Maximize" @click="emit('toggleMaximize')">
        <Square :size="11" :stroke-width="1.7" />
      </button>
      <button type="button" class="window-control close" title="Close" @click="emit('close')">
        <X :size="14" :stroke-width="1.7" />
      </button>
    </div>
  </header>
</template>
