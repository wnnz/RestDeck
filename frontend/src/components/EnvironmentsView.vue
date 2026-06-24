<script setup lang="ts">
import { CheckCircle2, Loader2, Plus, Trash2, Wand2, XCircle } from 'lucide-vue-next'
import { reactive } from 'vue'
import { domain } from '../../wailsjs/go/models'
import type { Translation } from '../i18n/messages'
import type { VariableSuggestion } from '../types'
import VoltSelect from './volt/VoltSelect.vue'
import VariableSuggestInput from './VariableSuggestInput.vue'
import VoltButton from './volt/VoltButton.vue'
import VoltCheckbox from './volt/VoltCheckbox.vue'
import VoltInputText from './volt/VoltInputText.vue'

const props = defineProps<{
  t: Translation
  mode: 'environment' | 'globals'
  collections: domain.Collection[]
  variableSuggestions: VariableSuggestion[]
  testJsonPath: (variable: domain.KeyValue) => Promise<string>
}>()

const envDraft = defineModel<{ id: string; name: string; variables: domain.KeyValue[] }>('envDraft', { required: true })
const globalsDraft = defineModel<domain.KeyValue[]>('globalsDraft', { required: true })

const emit = defineEmits<{
  addVariable: [target: domain.KeyValue[]]
  removeRow: [target: domain.KeyValue[], index: number]
}>()

const jsonPathTestState = reactive<Record<string, { busy: boolean; value: string; error: string }>>({})

function stateKey(variable: domain.KeyValue) {
  return variable.id || variable.key || 'jsonpath'
}

function testState(variable: domain.KeyValue) {
  const key = stateKey(variable)
  if (!jsonPathTestState[key]) jsonPathTestState[key] = { busy: false, value: '', error: '' }
  return jsonPathTestState[key]
}

async function testJSONPathVariable(variable: domain.KeyValue) {
  const state = testState(variable)
  state.busy = true
  state.value = ''
  state.error = ''
  try {
    state.value = await props.testJsonPath(variable)
  } catch (error) {
    state.error = error instanceof Error ? error.message : String(error)
  } finally {
    state.busy = false
  }
}

function requestOptions() {
  return (props.collections ?? []).flatMap((collection) => (collection.requests ?? []).map((request) => ({
    value: request.id,
    label: `${collection.name} / ${request.name}`
  })))
}

function valueTypeOptions() {
  return [
    { value: 'static', label: props.t.staticValue },
    { value: 'timestamp', label: props.t.timestampValue },
    { value: 'responseJsonPath', label: props.t.responseJsonPathValue }
  ]
}

function timestampOptions() {
  return [
    { value: 'seconds', label: props.t.timestampSeconds },
    { value: 'milliseconds', label: props.t.timestampMilliseconds },
    { value: 'iso', label: props.t.timestampIso }
  ]
}

function responseStrategyOptions() {
  return [
    { value: 'latestHistory', label: props.t.latestHistory },
    { value: 'alwaysRequest', label: props.t.alwaysRequest },
    { value: 'refreshAfter', label: props.t.refreshAfter }
  ]
}

</script>

<template>
  <div class="section-header">
    <div>
      <h2>{{ mode === 'globals' ? t.globals : t.environments }}</h2>
      <p v-if="mode === 'globals'">{{ t.localOnly }}</p>
    </div>
  </div>

  <section v-if="mode === 'environment'" class="environment-editor">
    <div class="kv-table spacious variable-table">
      <div class="kv-head variable-head"><span></span><span>{{ t.key }}</span><span>{{ t.valueType }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
      <div v-for="(variable, index) in envDraft.variables" :key="variable.id" class="kv-row variable-row">
        <VoltCheckbox v-model="variable.enabled" />
        <VoltInputText v-model="variable.key" />
        <VoltSelect v-model="variable.valueType" :options="valueTypeOptions()" />
        <div class="variable-value-cell">
          <VariableSuggestInput
            v-if="variable.valueType === 'static'"
            v-model="variable.value"
            :type="variable.secret ? 'password' : 'text'"
            :suggestions="variableSuggestions"
          />
          <VoltSelect v-else-if="variable.valueType === 'timestamp'" v-model="variable.timestampFormat" :options="timestampOptions()" />
          <div v-else class="response-var-grid">
            <VoltSelect v-model="variable.sourceRequestId" :options="[{ value: '', label: t.selectRequest }, ...requestOptions()]" />
            <VoltInputText v-model="variable.jsonPath" placeholder="$.items[0].id" />
            <VoltSelect v-model="variable.responseStrategy" :options="responseStrategyOptions()" />
            <VoltInputText v-if="variable.responseStrategy === 'refreshAfter'" v-model="variable.refreshAfterSeconds" type="number" />
            <VariableSuggestInput v-model="variable.fallbackValue" :suggestions="variableSuggestions" :placeholder="t.fallbackValue" />
            <VoltButton
              class="response-var-test"
              variant="secondary"
              :disabled="testState(variable).busy || !variable.sourceRequestId || !variable.jsonPath"
              @click="testJSONPathVariable(variable)"
            >
              <Loader2 v-if="testState(variable).busy" class="spin" :size="13" />
              <Wand2 v-else :size="13" />
              {{ t.testJsonPath }}
            </VoltButton>
            <div v-if="testState(variable).value || testState(variable).error" class="response-var-test-result">
              <CheckCircle2 v-if="testState(variable).value" :size="14" class="text-emerald-600" />
              <XCircle v-else :size="14" class="text-red-600" />
              <code>{{ testState(variable).value || testState(variable).error }}</code>
            </div>
          </div>
        </div>
        <VoltInputText v-model="variable.description" />
        <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('removeRow', envDraft.variables, index)"><Trash2 :size="13" /></VoltButton>
      </div>
      <VoltButton class="add-row" variant="secondary" @click="emit('addVariable', envDraft.variables)"><Plus :size="13" /> {{ t.addVariable }}</VoltButton>
    </div>
  </section>

  <div v-else class="kv-table spacious globals-table">
    <div class="kv-head"><span></span><span>{{ t.key }}</span><span>{{ t.value }}</span><span>{{ t.description }}</span><span></span></div>
    <div v-for="(variable, index) in globalsDraft" :key="variable.id" class="kv-row">
      <VoltCheckbox v-model="variable.enabled" />
      <VoltInputText v-model="variable.key" />
      <VariableSuggestInput v-model="variable.value" :suggestions="variableSuggestions" />
      <VoltInputText v-model="variable.description" />
      <VoltButton class="ghost-icon" size="icon" variant="ghost" @click="emit('removeRow', globalsDraft, index)"><Trash2 :size="13" /></VoltButton>
    </div>
    <VoltButton class="add-row" variant="secondary" @click="emit('addVariable', globalsDraft)"><Plus :size="13" /> {{ t.addGlobal }}</VoltButton>
  </div>
</template>
