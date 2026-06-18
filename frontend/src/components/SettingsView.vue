<script setup lang="ts">
import { KeyRound } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { Language } from '../types'

defineProps<{
  t: Translation
  theme: string
}>()

const language = defineModel<Language>('language', { required: true })
const emit = defineEmits<{
  'update:theme': [value: string]
}>()
</script>

<template>
  <div class="section-header"><div><h2>{{ t.settings }}</h2><p>{{ t.settingsHelp }}</p></div></div>
  <div class="settings-grid">
    <label><span>{{ t.language }}</span><select v-model="language"><option value="zh-CN">中文</option><option value="en-US">English</option></select></label>
    <label><span>{{ t.theme }}</span><select :value="theme" @change="emit('update:theme', ($event.target as HTMLSelectElement).value)"><option value="light">{{ t.light }}</option></select></label>
    <div class="settings-note">
      <KeyRound :size="16" />
      {{ t.encryptedNote }}
    </div>
  </div>
</template>
