<script setup lang="ts">
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { Language } from '../types'
import VoltSelect from './volt/VoltSelect.vue'
import VariableSuggestInput from './VariableSuggestInput.vue'

defineProps<{
  t: Translation
}>()

const language = defineModel<Language>('language', { required: true })
const settingsDraft = defineModel<domain.Settings>('settingsDraft', { required: true })
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.settings }}</h2></div>
  </div>
  <div class="settings-sections">
    <section class="settings-group">
      <div class="settings-group-title">{{ t.appSettings }}</div>
      <div class="settings-fields">
        <label class="settings-field">
          <span>{{ t.language }}</span>
          <VoltSelect v-model="language" :options="[{ value: 'zh-CN', label: '中文' }, { value: 'en-US', label: 'English' }]" />
        </label>
      </div>
    </section>

    <section class="settings-group">
      <div class="settings-group-title">{{ t.proxySettings }}</div>
      <div class="settings-fields">
        <label class="settings-field">
          <span>{{ t.defaultProxy }}</span>
          <VoltSelect v-model="settingsDraft.defaultProxy.mode" :options="[{ value: 'none', label: t.proxyNone }, { value: 'custom', label: t.proxyCustom }]" />
        </label>
        <label v-if="settingsDraft.defaultProxy.mode === 'custom'" class="settings-field">
          <span>{{ t.proxyUrl }}</span>
          <VariableSuggestInput v-model="settingsDraft.defaultProxy.url" placeholder="http://127.0.0.1:7890" :suggestions="[]" />
        </label>
        <label v-if="settingsDraft.defaultProxy.mode === 'custom'" class="settings-field">
          <span>{{ t.proxyNoProxy }}</span>
          <VariableSuggestInput v-model="settingsDraft.defaultProxy.noProxy" placeholder="localhost,127.0.0.1" :suggestions="[]" />
        </label>
      </div>
    </section>
  </div>
</template>
