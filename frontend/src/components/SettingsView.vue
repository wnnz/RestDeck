<script setup lang="ts">
import { KeyRound } from 'lucide-vue-next'
import type { Translation } from '../i18n/messages'
import type { Language } from '../types'

defineProps<{
  t: Translation
}>()

const language = defineModel<Language>('language', { required: true })
</script>

<template>
  <div class="section-header"><div><h2>{{ t.settings }}</h2><p>{{ t.settingsHelp }}</p></div></div>
  <div class="settings-grid">
    <label><span>{{ t.language }}</span><select v-model="language"><option value="zh-CN">中文</option><option value="en-US">English</option></select></label>
    <div class="settings-note">
      <KeyRound :size="16" />
      {{ t.encryptedNote }}
    </div>
  </div>
</template>
