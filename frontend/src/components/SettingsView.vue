<script setup lang="ts">
import { Trash2 } from 'lucide-vue-next'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { Language } from '../types'
import VoltSelect from './volt/VoltSelect.vue'
import VariableSuggestInput from './VariableSuggestInput.vue'
import VoltButton from './volt/VoltButton.vue'

defineProps<{
  t: Translation
  cookies: domain.Cookie[]
}>()

const language = defineModel<Language>('language', { required: true })
const settingsDraft = defineModel<domain.Settings>('settingsDraft', { required: true })

const emit = defineEmits<{
  deleteCookie: [cookie: domain.Cookie]
  clearCookies: []
}>()
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

    <section class="settings-group">
      <div class="settings-group-title">
        <span>{{ t.cookieJar }}</span>
        <VoltButton size="sm" variant="secondary" :disabled="!cookies.length" @click="emit('clearCookies')">{{ t.clearCookies }}</VoltButton>
      </div>
      <div class="kv-table spacious cookie-table">
        <div class="kv-head cookie-head"><span>{{ t.key }}</span><span>Domain</span><span>Path</span><span>{{ t.value }}</span><span></span></div>
        <div v-for="cookie in cookies" :key="`${cookie.domain}:${cookie.path}:${cookie.name}`" class="kv-row cookie-row">
          <code>{{ cookie.name }}</code>
          <span>{{ cookie.domain }}</span>
          <span>{{ cookie.path || '/' }}</span>
          <code>{{ cookie.value }}</code>
          <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('deleteCookie', cookie)"><Trash2 :size="13" /></VoltButton>
        </div>
        <div v-if="!cookies.length" class="empty-panel">{{ t.noCookies }}</div>
      </div>
    </section>
  </div>
</template>
