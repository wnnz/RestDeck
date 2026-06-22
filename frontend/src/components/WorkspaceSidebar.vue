<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Activity, ChevronDown, CircleAlert, Download, Globe2, Import, Loader2, MoreHorizontal, Pencil, Play, Plus, Radio, Trash2, X } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { NavKey } from '../types'

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
  environmentPanel: 'environment' | 'globals'
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
  createEnvironment: []
  selectEnvironment: [id: string]
  selectGlobalEnvironment: []
  renameEnvironment: [environment: domain.Environment, name: string]
  deleteEnvironment: [id: string]
  runCollection: []
}>()

const environmentMenuId = ref('')
const editingEnvironmentId = ref('')
const editingEnvironmentName = ref('')
const environmentRenameInput = ref<HTMLInputElement | null>(null)

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

function closeEnvironmentMenu() {
  environmentMenuId.value = ''
}

function openEnvironmentMenu(event: MouseEvent, environment: domain.Environment) {
  event.preventDefault()
  event.stopPropagation()
  environmentMenuId.value = environment.id
}

function selectEnvironment(id: string) {
  closeEnvironmentMenu()
  emit('selectEnvironment', id)
}

function selectGlobalEnvironment() {
  closeEnvironmentMenu()
  cancelEditingEnvironment()
  emit('selectGlobalEnvironment')
}

function startEditingEnvironment(environment: domain.Environment) {
  closeEnvironmentMenu()
  editingEnvironmentId.value = environment.id
  editingEnvironmentName.value = environment.name
  emit('selectEnvironment', environment.id)
  void nextTick(() => {
    environmentRenameInput.value?.focus()
    environmentRenameInput.value?.select()
  })
}

function setEnvironmentRenameInput(element: unknown) {
  environmentRenameInput.value = element instanceof HTMLInputElement ? element : null
}

function cancelEditingEnvironment() {
  editingEnvironmentId.value = ''
  editingEnvironmentName.value = ''
}

function saveEditingEnvironment(environment: domain.Environment) {
  if (editingEnvironmentId.value !== environment.id) return
  const nextName = editingEnvironmentName.value.trim()
  cancelEditingEnvironment()
  if (!nextName || nextName === environment.name) return
  emit('renameEnvironment', environment, nextName)
}

function deleteEnvironment(id: string) {
  closeEnvironmentMenu()
  if (editingEnvironmentId.value === id) cancelEditingEnvironment()
  emit('deleteEnvironment', id)
}

function handleDocumentClick(event: MouseEvent) {
  const target = event.target
  if (target instanceof Element && target.closest('.environment-row-wrap')) return
  closeEnvironmentMenu()
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
})
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
    <template v-else-if="activeNav === 'environments'">
      <div class="collection-toolbar environment-toolbar">
        <span class="sidebar-toolbar-title">{{ navLabel }}</span>
        <button class="icon-btn" type="button" :title="t.newEnvironment" @click="emit('createEnvironment')">
          <Plus :size="15" />
        </button>
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
          :class="['request-row', 'environment-row', 'global-environment-row', { active: environmentPanel === 'globals' }]"
          type="button"
          @click="selectGlobalEnvironment"
        >
          <Globe2 :size="14" />
          <span class="truncate">{{ t.globals }}</span>
          <span class="environment-special-badge">{{ t.globalScope }}</span>
        </button>
        <div
          v-for="env in environments"
          :key="env.id"
          class="environment-row-wrap"
        >
          <div v-if="editingEnvironmentId === env.id" :class="['request-row', 'environment-row', 'environment-row-editing', { active: environmentPanel === 'environment' && env.id === activeEnvironment?.id }]">
            <Globe2 :size="14" />
            <input
              :ref="setEnvironmentRenameInput"
              v-model="editingEnvironmentName"
              class="environment-rename-input"
              :placeholder="t.environmentName"
              @click.stop
              @keydown.enter.prevent="saveEditingEnvironment(env)"
              @keydown.esc.prevent="cancelEditingEnvironment"
              @blur="saveEditingEnvironment(env)"
            />
          </div>
          <button
            v-else
            :class="['request-row', 'environment-row', { active: environmentPanel === 'environment' && env.id === activeEnvironment?.id }]"
            type="button"
            @click="selectEnvironment(env.id)"
            @contextmenu="openEnvironmentMenu($event, env)"
          >
            <Globe2 :size="14" />
            <span class="truncate">{{ env.name }}</span>
          </button>
          <div v-if="environmentMenuId === env.id" class="action-menu environment-menu right" @click.stop>
            <button type="button" @click="startEditingEnvironment(env)">
              <Pencil :size="14" />
              {{ t.rename }}
            </button>
            <button class="danger-menu-item" type="button" @click="deleteEnvironment(env.id)">
              <Trash2 :size="14" />
              {{ t.delete }}
            </button>
          </div>
        </div>
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
  </aside>
</template>
