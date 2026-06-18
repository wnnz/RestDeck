<script setup lang="ts">
import { Plus, Save, Trash2 } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'

defineProps<{
  t: Translation
  activeEnvironment: domain.Environment | null
}>()

const envDraft = defineModel<{ id: string; name: string; variables: domain.KeyValue[] }>('envDraft', { required: true })
const globalsDraft = defineModel<domain.KeyValue[]>('globalsDraft', { required: true })

const emit = defineEmits<{
  saveEnvironment: []
  saveGlobals: []
  addVariable: [target: domain.KeyValue[]]
  removeRow: [target: domain.KeyValue[], index: number]
}>()
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.environments }}</h2><p>{{ activeEnvironment?.name ?? t.environmentSelected }}</p></div>
    <button class="toolbar-btn" @click="emit('saveEnvironment')"><Save :size="14" /> {{ t.save }}</button>
  </div>
  <label class="stack-label"><span>{{ t.key }}</span><input v-model="envDraft.name" class="field" /></label>
  <div class="kv-table spacious">
    <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
    <div v-for="(variable, index) in envDraft.variables" :key="variable.id" class="kv-row">
      <input v-model="variable.enabled" type="checkbox" />
      <input v-model="variable.key" />
      <input v-model="variable.value" :type="variable.secret ? 'password' : 'text'" />
      <input v-model="variable.description" />
      <button class="ghost-icon" @click="emit('removeRow', envDraft.variables, index)"><Trash2 :size="13" /></button>
    </div>
    <button class="add-row" @click="emit('addVariable', envDraft.variables)"><Plus :size="13" /> {{ t.addVariable }}</button>
  </div>
  <div class="section-header narrow">
    <div><h2>Globals</h2><p>{{ t.localOnly }}</p></div>
    <button class="toolbar-btn" @click="emit('saveGlobals')"><Save :size="14" /> {{ t.save }}</button>
  </div>
  <div class="kv-table spacious">
    <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
    <div v-for="(variable, index) in globalsDraft" :key="variable.id" class="kv-row">
      <input v-model="variable.enabled" type="checkbox" />
      <input v-model="variable.key" />
      <input v-model="variable.value" />
      <input v-model="variable.description" />
      <button class="ghost-icon" @click="emit('removeRow', globalsDraft, index)"><Trash2 :size="13" /></button>
    </div>
    <button class="add-row" @click="emit('addVariable', globalsDraft)"><Plus :size="13" /> {{ t.addGlobal }}</button>
  </div>
</template>
