<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { Activity, ChevronDown, CircleAlert, Code2, Copy, Download, Globe2, Import, MoreHorizontal, Pencil, Pin, Plus, Radio, Trash2, X } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { NavKey } from '../types'
import VoltButton from './volt/VoltButton.vue'
import VoltInputText from './volt/VoltInputText.vue'
import VoltPopover from './volt/VoltPopover.vue'

const props = defineProps<{
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
  optionsMenuOpen: boolean
  editingCollectionId: string
  editingCollectionName: string
  pendingDeleteCollectionId: string
}>()

const emit = defineEmits<{
  'update:collectionPickerOpen': [value: boolean]
  'update:optionsMenuOpen': [value: boolean]
  'update:editingCollectionName': [value: string]
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
  openOpenAPIModal: []
  exportCollection: []
  exportOpenAPICollection: []
  selectRequest: [request: domain.Request]
  generateRequestCode: [request: domain.Request]
  exportRequest: [request: domain.Request]
  pinRequest: [request: domain.Request]
  duplicateRequest: [request: domain.Request]
  deleteRequest: [request: domain.Request]
  createEnvironment: []
  selectEnvironment: [id: string]
  selectGlobalEnvironment: []
  renameEnvironment: [environment: domain.Environment, name: string]
  deleteEnvironment: [id: string]
}>()

const environmentMenuId = ref('')
const requestMenuId = ref('')
const pendingDeleteRequestId = ref('')
const editingEnvironmentId = ref('')
const editingEnvironmentName = ref('')
const environmentRenameInput = ref<InstanceType<typeof VoltInputText> | null>(null)
const collectionPickerPopover = ref<InstanceType<typeof VoltPopover> | null>(null)
const optionsMenuPopover = ref<InstanceType<typeof VoltPopover> | null>(null)
const requestMenuPopover = ref<InstanceType<typeof VoltPopover> | InstanceType<typeof VoltPopover>[] | null>(null)
const environmentMenuPopover = ref<InstanceType<typeof VoltPopover> | InstanceType<typeof VoltPopover>[] | null>(null)
const requestMenuAnchor = ref<HTMLElement | null>(null)
const environmentMenuAnchor = ref<HTMLElement | null>(null)

function toggleCollectionPicker(event: Event) {
  collectionPickerPopover.value?.toggle(event)
  emit('update:optionsMenuOpen', false)
  optionsMenuPopover.value?.hide()
}

function toggleOptionsMenu(event: Event) {
  optionsMenuPopover.value?.toggle(event)
  emit('update:collectionPickerOpen', false)
  collectionPickerPopover.value?.hide()
}

function closeEnvironmentMenu() {
  environmentMenuId.value = ''
  currentEnvironmentMenuPopover()?.hide()
}

function closeRequestMenu(resetDelete = true) {
  requestMenuId.value = ''
  if (resetDelete) pendingDeleteRequestId.value = ''
  currentRequestMenuPopover()?.hide()
}

function handleRequestMenuHide() {
  requestMenuId.value = ''
  pendingDeleteRequestId.value = ''
}

function closeActionMenus() {
  collectionPickerPopover.value?.hide()
  optionsMenuPopover.value?.hide()
  emit('update:collectionPickerOpen', false)
  emit('update:optionsMenuOpen', false)
}

function selectCollection(collection: domain.Collection) {
  collectionPickerPopover.value?.hide()
  emit('selectCollection', collection)
}

function createCollection() {
  collectionPickerPopover.value?.hide()
  emit('createCollection')
}

function createRequest() {
  closeActionMenus()
  emit('createRequest')
}

function openFetchModal() {
  closeActionMenus()
  emit('openFetchModal')
}

function openCurlModal() {
  closeActionMenus()
  emit('openCurlModal')
}

function openPostmanModal() {
  closeActionMenus()
  emit('openPostmanModal')
}

function openOpenAPIModal() {
  closeActionMenus()
  emit('openOpenAPIModal')
}

function exportCollection() {
  closeActionMenus()
  emit('exportCollection')
}

function exportOpenAPICollection() {
  closeActionMenus()
  emit('exportOpenAPICollection')
}

function openEnvironmentMenu(event: MouseEvent, environment: domain.Environment) {
  event.preventDefault()
  event.stopPropagation()
  const target = environmentMenuAnchor.value ?? event.currentTarget as HTMLElement
  if (environmentMenuAnchor.value) {
    environmentMenuAnchor.value.style.left = `${event.clientX + 6}px`
    environmentMenuAnchor.value.style.top = `${event.clientY + 6}px`
  }
  environmentMenuId.value = environment.id
  void nextTick(() => currentEnvironmentMenuPopover()?.show(event, target))
}

function openRequestMenu(event: MouseEvent, request: domain.Request) {
  event.preventDefault()
  event.stopPropagation()
  const target = requestMenuAnchor.value ?? event.currentTarget as HTMLElement
  if (requestMenuAnchor.value) {
    requestMenuAnchor.value.style.left = `${event.clientX + 6}px`
    requestMenuAnchor.value.style.top = `${event.clientY + 6}px`
  }
  pendingDeleteRequestId.value = ''
  requestMenuId.value = request.id
  void nextTick(() => currentRequestMenuPopover()?.show(event, target))
}

function currentRequestMenuPopover() {
  return Array.isArray(requestMenuPopover.value)
    ? requestMenuPopover.value[0]
    : requestMenuPopover.value
}

function currentEnvironmentMenuPopover() {
  return Array.isArray(environmentMenuPopover.value)
    ? environmentMenuPopover.value[0]
    : environmentMenuPopover.value
}

function generateRequestCode(request: domain.Request) {
  closeRequestMenu()
  emit('generateRequestCode', request)
}

function exportRequest(request: domain.Request) {
  closeRequestMenu()
  emit('exportRequest', request)
}

function pinRequest(request: domain.Request) {
  closeRequestMenu()
  emit('pinRequest', request)
}

function duplicateRequest(request: domain.Request) {
  closeRequestMenu()
  emit('duplicateRequest', request)
}

function deleteRequest(request: domain.Request) {
  if (pendingDeleteRequestId.value !== request.id) {
    pendingDeleteRequestId.value = request.id
    return
  }
  closeRequestMenu()
  emit('deleteRequest', request)
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

</script>

<template>
  <aside class="sidebar">
    <template v-if="activeNav === 'collections'">
      <div class="collection-toolbar">
        <div class="collection-picker-wrap">
          <VoltButton class="collection-link" variant="ghost" @click="toggleCollectionPicker">
            <span class="truncate">{{ activeCollection?.name ?? t.collections }}</span>
            <ChevronDown :size="13" />
          </VoltButton>
          <VoltPopover
            ref="collectionPickerPopover"
            class="collection-dropdown-popover"
            content-class="collection-dropdown"
            @hide="emit('update:collectionPickerOpen', false)"
            @show="emit('update:collectionPickerOpen', true)"
          >
            <div class="collection-dropdown-list">
              <div
                v-for="collection in collections"
                :key="collection.id"
                :class="['collection-option', { active: collection.id === activeCollection?.id }]"
              >
                <VoltInputText
                  v-if="editingCollectionId === collection.id"
                  :model-value="editingCollectionName"
                  class="collection-rename-input"
                  :placeholder="t.collectionName"
                  @update:model-value="emit('update:editingCollectionName', String($event))"
                  @keydown.enter="emit('saveEditingCollection', collection)"
                  @keydown.esc="emit('cancelEditingCollection')"
                  @blur="emit('saveEditingCollection', collection)"
                />
                <VoltButton v-else class="collection-option-name" variant="ghost" @click="selectCollection(collection)">
                  <span class="truncate">{{ collection.name }}</span>
                </VoltButton>
                <VoltButton class="ghost-icon" size="icon" variant="ghost" :title="t.editCollection" @click.stop="emit('startEditingCollection', collection)">
                  <Pencil :size="13" />
                </VoltButton>
                <VoltButton
                  :class="['ghost-icon', 'danger-icon', { pending: pendingDeleteCollectionId === collection.id }]"
                  size="icon"
                  variant="ghost"
                  :title="pendingDeleteCollectionId === collection.id ? t.confirmDeleteCollection : t.deleteCollection"
                  @click.stop="emit('deleteCollection', collection)"
                >
                  <CircleAlert v-if="pendingDeleteCollectionId === collection.id" :size="13" />
                  <X v-else :size="13" />
                </VoltButton>
              </div>
            </div>
            <VoltButton class="collection-new-option" variant="ghost" @click="createCollection">
              <Plus :size="14" />
              {{ t.newCollection }}
            </VoltButton>
          </VoltPopover>
        </div>

        <div class="collection-actions">
          <div class="menu-wrap">
            <VoltButton class="icon-btn" size="icon" variant="ghost" :title="t.collectionOptions" @click="toggleOptionsMenu">
              <MoreHorizontal :size="14" />
            </VoltButton>
            <VoltPopover
              ref="optionsMenuPopover"
              content-class="action-menu"
              @hide="emit('update:optionsMenuOpen', false)"
              @show="emit('update:optionsMenuOpen', true)"
            >
              <VoltButton variant="ghost" @click="openPostmanModal">
                <Import :size="14" />
                {{ t.importFromPostman }}
              </VoltButton>
              <VoltButton variant="ghost" @click="openOpenAPIModal">
                <Import :size="14" />
                {{ t.importFromOpenAPI }}
              </VoltButton>
              <VoltButton variant="ghost" @click="openFetchModal">
                <Import :size="14" />
                {{ t.importFromFetch }}
              </VoltButton>
              <VoltButton variant="ghost" @click="openCurlModal">
                <Import :size="14" />
                {{ t.importFromCurl }}
              </VoltButton>
              <VoltButton variant="ghost" :disabled="!activeCollection" @click="exportCollection">
                <Download :size="14" />
                Postman {{ t.export }}
              </VoltButton>
              <VoltButton variant="ghost" :disabled="!activeCollection" @click="exportOpenAPICollection">
                <Download :size="14" />
                OpenAPI {{ t.export }}
              </VoltButton>
            </VoltPopover>
          </div>
        </div>
      </div>
    </template>
    <template v-else-if="activeNav === 'environments'">
      <div class="collection-toolbar environment-toolbar">
        <span class="sidebar-toolbar-title">{{ navLabel }}</span>
        <VoltButton class="icon-btn" size="icon" variant="ghost" :title="t.newEnvironment" @click="emit('createEnvironment')">
          <Plus :size="15" />
        </VoltButton>
      </div>
    </template>
    <div v-else class="sidebar-title">
      <span>{{ navLabel }}</span>
    </div>

    <template v-if="activeNav === 'collections'">
      <div class="request-list">
        <span ref="requestMenuAnchor" class="context-menu-anchor" aria-hidden="true"></span>
        <VoltButton class="request-row new-request-row" variant="ghost" @click="createRequest">
          <Plus :size="14" />
          <span>{{ t.newRequest }}</span>
        </VoltButton>
        <VoltButton
          v-for="request in filteredRequests"
          :key="request.id"
          :class="['request-row', { active: request.id === activeRequest?.id }]"
          @click="emit('selectRequest', request)"
          @contextmenu="openRequestMenu($event, request)"
        >
          <span :class="['method', request.method.toLowerCase()]">{{ request.method }}</span>
          <span class="truncate">{{ request.name }}</span>
        </VoltButton>
        <VoltPopover
          v-if="requestMenuId"
          ref="requestMenuPopover"
          content-class="action-menu"
          @hide="handleRequestMenuHide"
        >
          <template v-for="request in filteredRequests" :key="request.id">
            <template v-if="requestMenuId === request.id">
              <VoltButton variant="ghost" @click="generateRequestCode(request)">
                <Code2 :size="14" />
                {{ t.generateCode }}
              </VoltButton>
              <VoltButton variant="ghost" @click="exportRequest(request)">
                <Download :size="14" />
                {{ t.export }}
              </VoltButton>
              <VoltButton variant="ghost" @click="pinRequest(request)">
                <Pin :size="14" />
                {{ t.pinToTop }}
              </VoltButton>
              <VoltButton variant="ghost" @click="duplicateRequest(request)">
                <Copy :size="14" />
                {{ t.duplicate }}
              </VoltButton>
              <VoltButton class="danger-menu-item" variant="ghost" @click="deleteRequest(request)">
                <CircleAlert v-if="pendingDeleteRequestId === request.id" :size="14" />
                <Trash2 v-else :size="14" />
                {{ pendingDeleteRequestId === request.id ? t.confirmDeleteRequest : t.delete }}
              </VoltButton>
            </template>
          </template>
        </VoltPopover>
        <div v-if="!activeCollection" class="side-note">{{ t.createOrSelect }}</div>
      </div>
    </template>

    <template v-else-if="activeNav === 'environments'">
      <div class="request-list">
        <span ref="environmentMenuAnchor" class="context-menu-anchor" aria-hidden="true"></span>
        <VoltButton
          :class="['request-row', 'environment-row', 'global-environment-row', { active: environmentPanel === 'globals' }]"
          variant="ghost"
          @click="selectGlobalEnvironment"
        >
          <Globe2 :size="14" />
          <span class="truncate">{{ t.globals }}</span>
          <span class="environment-special-badge">{{ t.globalScope }}</span>
        </VoltButton>
        <div
          v-for="env in environments"
          :key="env.id"
          class="environment-row-wrap"
        >
          <div v-if="editingEnvironmentId === env.id" :class="['request-row', 'environment-row', 'environment-row-editing', { active: environmentPanel === 'environment' && env.id === activeEnvironment?.id }]">
            <Globe2 :size="14" />
            <VoltInputText
              ref="environmentRenameInput"
              v-model="editingEnvironmentName"
              class="environment-rename-input"
              :placeholder="t.environmentName"
              @click.stop
              @keydown.enter.prevent="saveEditingEnvironment(env)"
              @keydown.esc.prevent="cancelEditingEnvironment"
              @blur="saveEditingEnvironment(env)"
            />
          </div>
          <VoltButton
            v-else
            :class="['request-row', 'environment-row', { active: environmentPanel === 'environment' && env.id === activeEnvironment?.id }]"
            variant="ghost"
            @click="selectEnvironment(env.id)"
            @contextmenu="openEnvironmentMenu($event, env)"
          >
            <Globe2 :size="14" />
            <span class="truncate">{{ env.name }}</span>
          </VoltButton>
          <VoltPopover
            v-if="environmentMenuId === env.id"
            ref="environmentMenuPopover"
            content-class="action-menu"
            @hide="environmentMenuId = ''"
          >
            <VoltButton variant="ghost" @click="startEditingEnvironment(env)">
              <Pencil :size="14" />
              {{ t.rename }}
            </VoltButton>
            <VoltButton class="danger-menu-item" variant="ghost" @click="deleteEnvironment(env.id)">
              <Trash2 :size="14" />
              {{ t.delete }}
            </VoltButton>
          </VoltPopover>
        </div>
      </div>
    </template>

    <template v-else-if="activeNav === 'realtime'">
      <div class="side-note">{{ t.realtimeHelp }}</div>
      <VoltButton class="request-row active" variant="ghost">
        <Radio :size="14" />
        <span>WebSocket</span>
      </VoltButton>
      <VoltButton class="request-row" variant="ghost">
        <Activity :size="14" />
        <span>SSE</span>
      </VoltButton>
    </template>
  </aside>
</template>
