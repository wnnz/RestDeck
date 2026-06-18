<script setup lang="ts">
import { Activity, ChevronDown, CircleAlert, Download, Globe2, Import, Loader2, MoreHorizontal, Pencil, Play, Plus, Radio, X } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { NavKey } from '../types'
import { statusClass } from '../utils/format'

defineProps<{
  t: Translation
  activeNav: NavKey
  navLabel: string
  collections: domain.Collection[]
  activeCollection: domain.Collection | null
  filteredRequests: domain.Request[]
  activeRequest: domain.Request | null
  environments: domain.Environment[]
  activeEnvironment: domain.Environment | null
  history: domain.HistoryItem[]
  collectionPickerOpen: boolean
  addMenuOpen: boolean
  optionsMenuOpen: boolean
  editingCollectionId: string
  editingCollectionName: string
  pendingDeleteCollectionId: string
  runnerBusy: boolean
}>()

const emit = defineEmits<{
  'update:collectionPickerOpen': [value: boolean]
  'update:addMenuOpen': [value: boolean]
  'update:optionsMenuOpen': [value: boolean]
  'update:editingCollectionName': [value: string]
  setToolbarRef: [element: HTMLElement | null]
  selectCollection: [collection: domain.Collection]
  startEditingCollection: [collection: domain.Collection]
  cancelEditingCollection: []
  saveEditingCollection: [collection: domain.Collection]
  deleteCollection: [collection: domain.Collection]
  createCollection: []
  createRequest: []
  openFetchModal: []
  openCurlModal: []
  openPostmanModal: []
  exportCollection: []
  selectRequest: [request: domain.Request]
  selectEnvironment: [id: string]
  runCollection: []
}>()

function toggleCollectionPicker(open: boolean) {
  emit('update:collectionPickerOpen', open)
  emit('update:addMenuOpen', false)
  emit('update:optionsMenuOpen', false)
}

function toggleAddMenu(open: boolean) {
  emit('update:addMenuOpen', open)
  emit('update:collectionPickerOpen', false)
  emit('update:optionsMenuOpen', false)
}

function toggleOptionsMenu(open: boolean) {
  emit('update:optionsMenuOpen', open)
  emit('update:collectionPickerOpen', false)
  emit('update:addMenuOpen', false)
}
</script>

<template>
  <aside class="sidebar">
    <template v-if="activeNav === 'collections'">
      <div :ref="(element) => emit('setToolbarRef', element as HTMLElement | null)" class="collection-toolbar">
        <div class="collection-picker-wrap">
          <button class="collection-link" type="button" @click="toggleCollectionPicker(!collectionPickerOpen)">
            <span class="truncate">{{ activeCollection?.name ?? t.collections }}</span>
            <ChevronDown :size="13" />
          </button>
          <div v-if="collectionPickerOpen" class="collection-dropdown">
            <div class="collection-dropdown-list">
              <div
                v-for="collection in collections"
                :key="collection.id"
                :class="['collection-option', { active: collection.id === activeCollection?.id }]"
              >
                <input
                  v-if="editingCollectionId === collection.id"
                  :value="editingCollectionName"
                  class="collection-rename-input"
                  :placeholder="t.collectionName"
                  @input="emit('update:editingCollectionName', ($event.target as HTMLInputElement).value)"
                  @keydown.enter="emit('saveEditingCollection', collection)"
                  @keydown.esc="emit('cancelEditingCollection')"
                  @blur="emit('saveEditingCollection', collection)"
                />
                <button v-else class="collection-option-name" type="button" @click="emit('selectCollection', collection)">
                  <span class="truncate">{{ collection.name }}</span>
                </button>
                <button class="ghost-icon" type="button" :title="t.editCollection" @click.stop="emit('startEditingCollection', collection)">
                  <Pencil :size="13" />
                </button>
                <button
                  :class="['ghost-icon', 'danger-icon', { pending: pendingDeleteCollectionId === collection.id }]"
                  type="button"
                  :title="pendingDeleteCollectionId === collection.id ? t.confirmDeleteCollection : t.deleteCollection"
                  @click.stop="emit('deleteCollection', collection)"
                >
                  <CircleAlert v-if="pendingDeleteCollectionId === collection.id" :size="13" />
                  <X v-else :size="13" />
                </button>
              </div>
            </div>
            <button class="collection-new-option" type="button" @click="emit('createCollection')">
              <Plus :size="14" />
              {{ t.newCollection }}
            </button>
          </div>
        </div>

        <div class="collection-actions">
          <div class="menu-wrap">
            <button class="icon-btn" type="button" :title="t.new" @click="toggleAddMenu(!addMenuOpen)">
              <Plus :size="15" />
            </button>
            <div v-if="addMenuOpen" class="action-menu">
              <button type="button" @click="emit('createRequest')">
                <Plus :size="14" />
                {{ t.newRequest }}
              </button>
              <button type="button" @click="emit('openFetchModal')">
                <Import :size="14" />
                {{ t.importFromFetch }}
              </button>
              <button type="button" @click="emit('openCurlModal')">
                <Import :size="14" />
                {{ t.importFromCurl }}
              </button>
            </div>
          </div>
          <div class="menu-wrap">
            <button class="icon-btn" type="button" :title="t.collectionOptions" @click="toggleOptionsMenu(!optionsMenuOpen)">
              <MoreHorizontal :size="14" />
            </button>
            <div v-if="optionsMenuOpen" class="action-menu right">
              <button type="button" @click="emit('openPostmanModal')">
                <Import :size="14" />
                {{ t.importFromPostman }}
              </button>
              <button type="button" :disabled="!activeCollection" @click="emit('exportCollection')">
                <Download :size="14" />
                {{ t.export }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
    <div v-else class="sidebar-title">
      <span>{{ navLabel }}</span>
    </div>

    <template v-if="activeNav === 'collections'">
      <div class="request-list">
        <button
          v-for="request in filteredRequests"
          :key="request.id"
          :class="['request-row', { active: request.id === activeRequest?.id }]"
          @click="emit('selectRequest', request)"
        >
          <span :class="['method', request.method.toLowerCase()]">{{ request.method }}</span>
          <span class="truncate">{{ request.name }}</span>
        </button>
        <div v-if="!activeCollection" class="side-note">{{ t.createOrSelect }}</div>
      </div>
    </template>

    <template v-else-if="activeNav === 'environments'">
      <div class="request-list">
        <button
          v-for="env in environments"
          :key="env.id"
          :class="['request-row', { active: env.id === activeEnvironment?.id }]"
          @click="emit('selectEnvironment', env.id)"
        >
          <Globe2 :size="14" />
          <span class="truncate">{{ env.name }}</span>
        </button>
      </div>
    </template>

    <template v-else-if="activeNav === 'history'">
      <div class="request-list">
        <button v-for="item in history" :key="item.id" class="request-row" @click="emit('selectRequest', item.request)">
          <span :class="['method', item.method.toLowerCase()]">{{ item.method }}</span>
          <span class="truncate">{{ item.name || item.url }}</span>
          <span :class="['history-code', statusClass(item.statusCode)]">{{ item.statusCode || '-' }}</span>
        </button>
      </div>
    </template>

    <template v-else-if="activeNav === 'runner'">
      <div class="side-note">{{ t.runnerHelp }}</div>
      <button class="primary-wide" :disabled="runnerBusy || !activeCollection" @click="emit('runCollection')">
        <Loader2 v-if="runnerBusy" class="spin" :size="14" />
        <Play v-else :size="14" />
        Run collection
      </button>
    </template>

    <template v-else-if="activeNav === 'realtime'">
      <div class="side-note">{{ t.realtimeHelp }}</div>
      <button class="request-row active" type="button">
        <Radio :size="14" />
        <span>WebSocket</span>
      </button>
      <button class="request-row" type="button">
        <Activity :size="14" />
        <span>SSE</span>
      </button>
    </template>

    <template v-else>
      <div class="side-note">{{ t.settingsSide }}</div>
    </template>
  </aside>
</template>
