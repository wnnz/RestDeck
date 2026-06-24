<script setup lang="ts">
import { CheckCircle2, Loader2, Plus, Trash2, Wand2, XCircle } from 'lucide-vue-next'
import { computed, reactive, ref, watch } from 'vue'
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

const selectedEnvironmentVariableId = ref('')
const selectedGlobalVariableId = ref('')
const jsonPathTestState = reactive<Record<string, { busy: boolean; value: string; error: string }>>({})

const activeVariables = computed(() => props.mode === 'globals' ? globalsDraft.value : envDraft.value.variables)
const selectedVariableIndex = computed(() => activeVariables.value.findIndex((variable) => variableId(variable) === selectedVariableId()))
const selectedVariable = computed(() => selectedVariableIndex.value >= 0 ? activeVariables.value[selectedVariableIndex.value] : null)

watch(() => [props.mode, ...activeVariables.value.map((variable, index) => variableId(variable) || String(index))], () => {
  ensureSelectedVariable()
}, { immediate: true })

function stateKey(variable: domain.KeyValue) {
  return variable.id || variable.key || 'jsonpath'
}

function variableId(variable: domain.KeyValue) {
  return variable.id || variable.key
}

function selectedVariableId() {
  return props.mode === 'globals' ? selectedGlobalVariableId.value : selectedEnvironmentVariableId.value
}

function setSelectedVariableId(id: string) {
  if (props.mode === 'globals') {
    selectedGlobalVariableId.value = id
    return
  }
  selectedEnvironmentVariableId.value = id
}

function ensureSelectedVariable() {
  const variables = activeVariables.value
  if (!variables.length) {
    setSelectedVariableId('')
    return
  }
  if (!variables.some((variable) => variableId(variable) === selectedVariableId())) {
    setSelectedVariableId(variableId(variables[0]))
  }
}

function selectVariable(variable: domain.KeyValue) {
  setSelectedVariableId(variableId(variable))
}

function addVariableAndSelect(target: domain.KeyValue[]) {
  const nextIndex = target.length
  emit('addVariable', target)
  const variable = target[nextIndex] ?? target[target.length - 1]
  if (variable) selectVariable(variable)
}

function removeVariable(target: domain.KeyValue[], index: number) {
  const removingActive = target[index] && variableId(target[index]) === selectedVariableId()
  emit('removeRow', target, index)
  if (!removingActive) return
  const next = target[Math.min(index, target.length - 1)] ?? target[index - 1]
  setSelectedVariableId(next ? variableId(next) : '')
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

function optionLabel(options: Array<{ value: string; label: string }>, value: string, fallback = '') {
  return options.find((option) => option.value === value)?.label ?? fallback
}

function variableName(variable: domain.KeyValue) {
  return variable.key?.trim() || props.t.unnamedVariable
}

function variableTypeLabel(variable: domain.KeyValue) {
  if (props.mode === 'globals') return props.t.staticValue
  return optionLabel(valueTypeOptions(), variable.valueType, props.t.staticValue)
}

function variablePreview(variable: domain.KeyValue) {
  if (props.mode === 'globals' || variable.valueType === 'static') {
    if (variable.secret && variable.value) return '••••••'
    return variable.value || props.t.emptyValue
  }
  if (variable.valueType === 'timestamp') {
    return optionLabel(timestampOptions(), variable.timestampFormat, props.t.timestampSeconds)
  }
  return variable.jsonPath || props.t.responseConfig
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

  <section class="environment-editor">
    <div class="variable-workbench">
      <aside class="variable-list-panel">
        <div class="variable-list-header">
          <strong>{{ mode === 'globals' ? t.globals : t.environmentVariable }}</strong>
          <VoltButton size="icon" variant="secondary" @click="addVariableAndSelect(activeVariables)"><Plus :size="14" /></VoltButton>
        </div>

        <div class="variable-list">
          <button
            v-for="(variable, index) in activeVariables"
            :key="variable.id || index"
            class="variable-list-item"
            :class="{ active: variableId(variable) === selectedVariableId(), disabled: !variable.enabled }"
            type="button"
            @click="selectVariable(variable)"
          >
            <VoltCheckbox v-model="variable.enabled" @click.stop />
            <span class="variable-list-content">
              <strong>{{ variableName(variable) }}</strong>
              <span>{{ variableTypeLabel(variable) }} · {{ variablePreview(variable) }}</span>
            </span>
          </button>

          <div v-if="!activeVariables.length" class="variable-list-empty">{{ t.noVariables }}</div>
        </div>
      </aside>

      <section class="variable-detail-panel">
        <template v-if="selectedVariable">
          <div class="variable-detail-header">
            <div>
              <span>{{ t.variableDetail }}</span>
              <strong>{{ variableName(selectedVariable) }}</strong>
            </div>
            <VoltButton size="icon" variant="ghost" @click="removeVariable(activeVariables, selectedVariableIndex)"><Trash2 :size="14" /></VoltButton>
          </div>

          <div class="variable-detail-grid">
            <label class="variable-detail-field">
              <span>{{ t.enabled }}</span>
              <span class="variable-toggle">
                <VoltCheckbox v-model="selectedVariable.enabled" />
                {{ selectedVariable.enabled ? t.enabled : t.disabled }}
              </span>
            </label>
            <label class="variable-detail-field">
              <span>{{ t.key }}</span>
              <VoltInputText v-model="selectedVariable.key" />
            </label>
            <label v-if="mode === 'environment'" class="variable-detail-field">
              <span>{{ t.valueType }}</span>
              <VoltSelect v-model="selectedVariable.valueType" :options="valueTypeOptions()" />
            </label>
            <label
              v-if="mode === 'globals' || selectedVariable.valueType === 'static'"
              class="variable-detail-field variable-detail-wide"
            >
              <span>{{ t.value }}</span>
              <VariableSuggestInput
                v-model="selectedVariable.value"
                :type="selectedVariable.secret ? 'password' : 'text'"
                :suggestions="variableSuggestions"
              />
            </label>
            <label v-else-if="selectedVariable.valueType === 'timestamp'" class="variable-detail-field variable-detail-wide">
              <span>{{ t.timestampValue }}</span>
              <VoltSelect v-model="selectedVariable.timestampFormat" :options="timestampOptions()" />
            </label>
            <label class="variable-detail-field variable-detail-wide">
              <span>{{ t.description }}</span>
              <VoltInputText v-model="selectedVariable.description" />
            </label>
          </div>

          <div v-if="mode === 'environment' && selectedVariable.valueType === 'responseJsonPath'" class="variable-response-card">
            <div class="variable-response-title">{{ t.responseConfig }}</div>
            <div class="variable-response-grid">
              <label class="variable-detail-field variable-detail-wide">
                <span>{{ t.sourceRequest }}</span>
                <VoltSelect v-model="selectedVariable.sourceRequestId" :options="[{ value: '', label: t.selectRequest }, ...requestOptions()]" />
              </label>
              <label class="variable-detail-field">
                <span>{{ t.jsonPath }}</span>
                <VoltInputText v-model="selectedVariable.jsonPath" placeholder="$.items[0].id" />
              </label>
              <label class="variable-detail-field">
                <span>{{ t.readStrategy }}</span>
                <VoltSelect v-model="selectedVariable.responseStrategy" :options="responseStrategyOptions()" />
              </label>
              <label v-if="selectedVariable.responseStrategy === 'refreshAfter'" class="variable-detail-field">
                <span>{{ t.refreshSeconds }}</span>
                <VoltInputText v-model="selectedVariable.refreshAfterSeconds" type="number" />
              </label>
              <label class="variable-detail-field variable-detail-wide">
                <span>{{ t.fallbackValue }}</span>
                <VariableSuggestInput v-model="selectedVariable.fallbackValue" :suggestions="variableSuggestions" :placeholder="t.fallbackValue" />
              </label>
            </div>

            <div class="response-var-actions">
              <VoltButton
                class="response-var-test"
                variant="secondary"
                :disabled="testState(selectedVariable).busy || !selectedVariable.sourceRequestId || !selectedVariable.jsonPath"
                @click="testJSONPathVariable(selectedVariable)"
              >
                <Loader2 v-if="testState(selectedVariable).busy" class="spin" :size="13" />
                <Wand2 v-else :size="13" />
                {{ t.testJsonPath }}
              </VoltButton>
              <label v-if="testState(selectedVariable).value || testState(selectedVariable).error" class="variable-detail-field response-var-result-field">
                <span>{{ t.testResult }}</span>
                <div class="response-var-test-result">
                  <CheckCircle2 v-if="testState(selectedVariable).value" :size="14" class="text-emerald-600" />
                  <XCircle v-else :size="14" class="text-red-600" />
                  <VoltInputText
                    :model-value="testState(selectedVariable).value || testState(selectedVariable).error"
                    input-class="response-var-test-output"
                    readonly
                  />
                </div>
              </label>
            </div>
          </div>
        </template>

        <div v-else class="variable-detail-empty">
          <span>{{ t.selectVariableHint }}</span>
          <VoltButton variant="secondary" @click="addVariableAndSelect(activeVariables)"><Plus :size="14" /> {{ mode === 'globals' ? t.addGlobal : t.addVariable }}</VoltButton>
        </div>
      </section>
    </div>
  </section>
</template>
