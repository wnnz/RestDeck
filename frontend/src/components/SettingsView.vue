<script setup lang="ts">
import { KeyRound } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { Language } from '../types'
import VariableSuggestInput from './VariableSuggestInput.vue'

defineProps<{
  t: Translation
}>()

const language = defineModel<Language>('language', { required: true })
const settingsDraft = defineModel<domain.Settings>('settingsDraft', { required: true })

const emit = defineEmits<{
  saveSettings: []
}>()
</script>

<template>
  <div class="section-header">
    <div><h2>{{ t.settings }}</h2><p>{{ t.settingsHelp }}</p></div>
    <button class="toolbar-btn" @click="emit('saveSettings')">{{ t.save }}</button>
  </div>
  <div class="settings-sections">
    <section class="settings-group">
      <div class="settings-group-title">{{ t.appSettings }}</div>
      <div class="settings-fields">
        <label class="settings-field">
          <span>{{ t.language }}</span>
          <select v-model="language"><option value="zh-CN">中文</option><option value="en-US">English</option></select>
        </label>
      </div>
    </section>

    <section class="settings-group">
      <div class="settings-group-title">{{ t.proxySettings }}</div>
      <div class="settings-fields">
        <label class="settings-field">
          <span>{{ t.defaultProxy }}</span>
          <select v-model="settingsDraft.defaultProxy.mode"><option value="none">{{ t.proxyNone }}</option><option value="custom">{{ t.proxyCustom }}</option></select>
        </label>
        <label v-if="settingsDraft.defaultProxy.mode === 'custom'" class="settings-field">
          <span>{{ t.proxyUrl }}</span>
          <VariableSuggestInput v-model="settingsDraft.defaultProxy.url" placeholder="http://127.0.0.1:7890" :suggestions="[]" />
        </label>
      </div>
    </section>

    <section class="settings-group">
      <div class="settings-group-title">{{ t.securitySettings }}</div>
      <div class="settings-note">
        <KeyRound :size="16" />
        {{ t.encryptedNote }}
      </div>
    </section>
  </div>
</template>
