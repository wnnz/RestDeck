import { ref, type ComputedRef, type Ref } from 'vue'
import { domain } from '../../wailsjs/go/models'
import type { RunnerFailurePolicy, RunnerQueueItem } from '../types'
import { formatError } from '../utils/format'
import { cloneRequest, normalizeRequest, syncFormBody } from '../utils/requestModel'
import { buildRunnerQueue, markPendingRunnerItems, newRunnerResult, patchRunnerQueueItem, summarizeRunnerResponse } from '../utils/runnerModel'

type RunnerLabels = {
  runner: string
  passed: string
  failed: string
  skipped: string
  stopped: string
  running: string
  sendingRequest: string
  tests: string
}

type UseRunnerControllerOptions = {
  activeCollection: ComputedRef<domain.Collection | null>
  activeEnvironment: ComputedRef<domain.Environment | null>
  globalsDraft: Ref<domain.KeyValue[]>
  labels: ComputedRef<RunnerLabels>
  statusMessage: Ref<string>
  saveRequest: (request: domain.Request) => Promise<domain.WorkspaceState>
  sendRequest: (request: domain.Request, environmentId: string, globals: domain.KeyValue[]) => Promise<domain.Response>
  getState: () => Promise<domain.WorkspaceState>
  setState: (state: domain.WorkspaceState) => void
  saveRunnerResult: (result: domain.RunnerResult) => Promise<domain.WorkspaceState>
}

export function useRunnerController(options: UseRunnerControllerOptions) {
  const runnerBusy = ref(false)
  const runnerResult = ref<domain.RunnerResult | null>(null)
  const runnerScope = ref<'collection' | 'request'>('collection')
  const runnerIterations = ref(1)
  const runnerRequestId = ref('')
  const runnerQueue = ref<RunnerQueueItem[]>([])
  const runnerFailurePolicy = ref<RunnerFailurePolicy>('continue')
  const runnerStopRequested = ref(false)
  const runnerRetryCount = ref(0)
  const runnerDelayMs = ref(0)

  function ensureRunnerRequest(activeRequestId: string) {
    const collection = options.activeCollection.value
    if (!runnerRequestId.value || !collection?.requests?.some((request) => request.id === runnerRequestId.value)) {
      runnerRequestId.value = activeRequestId || collection?.requests?.[0]?.id || ''
    }
  }

  function setRunnerRequest(id: string) {
    runnerRequestId.value = id
  }

  function selectRunnerRequest(id: string) {
    setRunnerRequest(id)
    resetRunnerOutput()
  }

  function setRunnerScope(scope: 'collection' | 'request') {
    runnerScope.value = scope
    resetRunnerOutput()
  }

  function setRunnerIterations(iterations: number) {
    runnerIterations.value = Math.max(1, Math.floor(iterations || 1))
    resetRunnerOutput()
  }

  function setRunnerFailurePolicy(policy: RunnerFailurePolicy) {
    runnerFailurePolicy.value = policy === 'stopOnFailure' ? 'stopOnFailure' : 'continue'
    resetRunnerOutput()
  }

  function setRunnerRetryCount(value: number) {
    runnerRetryCount.value = Math.max(0, Math.floor(value || 0))
    resetRunnerOutput()
  }

  function setRunnerDelayMs(value: number) {
    runnerDelayMs.value = Math.max(0, Math.floor(value || 0))
    resetRunnerOutput()
  }

  function stopRunner() {
    if (!runnerBusy.value) return
    runnerStopRequested.value = true
    options.statusMessage.value = options.labels.value.stopped
  }

  function resetRunnerOutput() {
    runnerQueue.value = []
    runnerResult.value = null
    runnerStopRequested.value = false
  }

  async function runActiveCollection() {
    const collection = options.activeCollection.value
    if (!collection) return
    const iterations = Math.max(1, runnerIterations.value)
    const requests = collection.requests ?? []
    const startedAt = Date.now()
    runnerBusy.value = true
    runnerStopRequested.value = false
    try {
      runnerQueue.value = buildRunnerQueue(requests, iterations)
      runnerResult.value = newRunnerResult({
        collectionId: collection.id,
        environmentId: options.activeEnvironment.value?.id ?? '',
        name: collection.name,
        iterations
      })
      for (let iteration = 1; iteration <= iterations; iteration++) {
        for (const request of requests) {
          if (runnerStopRequested.value) {
            skipPendingRunnerItems(options.labels.value.stopped)
            break
          }
          const passed = await runRunnerQueueRequest(request, iteration)
          if (!passed && runnerFailurePolicy.value === 'stopOnFailure') {
            skipPendingRunnerItems(options.labels.value.failed)
            runnerStopRequested.value = true
            break
          }
        }
        if (runnerStopRequested.value) break
      }
      runnerResult.value.durationMs = Date.now() - startedAt
      await persistRunnerResult()
      options.statusMessage.value = runnerStatusMessage()
    } catch (error) {
      options.statusMessage.value = formatError(error)
    } finally {
      runnerBusy.value = false
    }
  }

  async function runRunnerRequest() {
    const request = options.activeCollection.value?.requests?.find((item) => item.id === runnerRequestId.value)
    if (!request) return
    const startedAt = Date.now()
    runnerBusy.value = true
    runnerStopRequested.value = false
    try {
      runnerQueue.value = buildRunnerQueue([request], 1)
      runnerResult.value = newRunnerResult({
        collectionId: request.collectionId,
        environmentId: options.activeEnvironment.value?.id ?? '',
        name: request.name || request.url,
        iterations: 1
      })
      await runRunnerQueueRequest(request, 1)
      runnerResult.value.durationMs = Date.now() - startedAt
      await persistRunnerResult()
      options.statusMessage.value = runnerStatusMessage()
    } catch (error) {
      options.statusMessage.value = formatError(error)
    } finally {
      runnerBusy.value = false
    }
  }

  async function runRunnerQueueRequest(request: domain.Request, iteration: number) {
    const queueId = `${request.id}-${iteration}`
    updateRunnerQueueItem(queueId, { status: 'running', message: options.labels.value.running })
    const startedAt = Date.now()
    const requestToSend = normalizeRequest(cloneRequest(request))
    try {
      syncFormBody(requestToSend)
      options.statusMessage.value = options.labels.value.sendingRequest
      const savedState = await options.saveRequest(requestToSend)
      options.setState(savedState)
      let result: domain.Response | null = null
      let lastError: unknown = null
      for (let attempt = 0; attempt <= runnerRetryCount.value; attempt++) {
        try {
          result = await options.sendRequest(requestToSend, options.activeEnvironment.value?.id ?? '', options.globalsDraft.value)
          const summary = summarizeRunnerResponse(result, options.labels.value.tests)
          if (summary.passed || attempt >= runnerRetryCount.value) break
          updateRunnerQueueItem(queueId, { message: `${summary.message} · retry ${attempt + 1}/${runnerRetryCount.value}` })
        } catch (error) {
          lastError = error
          if (attempt >= runnerRetryCount.value) throw error
          updateRunnerQueueItem(queueId, { message: `${formatError(error)} · retry ${attempt + 1}/${runnerRetryCount.value}` })
        }
      }
      if (!result) throw lastError ?? new Error('No response')
      const summary = summarizeRunnerResponse(result, options.labels.value.tests)
      updateRunnerQueueItem(queueId, {
        status: summary.passed ? 'passed' : 'failed',
        url: result.requestedUrl || requestToSend.url,
        statusCode: result.statusCode,
        durationMs: result.durationMs,
        message: summary.message,
        request: result.request,
        response: result,
        testResults: result.testResults ?? []
      })
      addRunnerResultItem(requestToSend, summary.passed, summary.message, result, iteration)
      try {
        const latestState = await options.getState()
        options.setState(latestState)
      } catch (error) {
        options.statusMessage.value = formatError(error)
      }
    } catch (error) {
      const message = formatError(error)
      updateRunnerQueueItem(queueId, {
        status: 'failed',
        durationMs: Date.now() - startedAt,
        message
      })
      addRunnerResultItem(requestToSend, false, message, null, iteration)
      return false
    }
    if (runnerDelayMs.value > 0) await delay(runnerDelayMs.value)
    return runnerQueue.value.find((item) => item.id === queueId)?.status === 'passed'
  }

  function skipPendingRunnerItems(message: string) {
    runnerQueue.value = markPendingRunnerItems(runnerQueue.value, {
      status: 'skipped',
      message
    })
  }

  function addRunnerResultItem(request: domain.Request, passed: boolean, message: string, response: domain.Response | null, iteration: number) {
    if (!runnerResult.value) return
    runnerResult.value.passed += passed ? 1 : 0
    runnerResult.value.failed += passed ? 0 : 1
    runnerResult.value.items = [
      ...(runnerResult.value.items ?? []),
      new domain.TestResult({ name: request.name || request.url, passed, message })
    ]
    runnerResult.value.details = [
      ...(runnerResult.value.details ?? []),
      new domain.RunnerRequestResult({
        id: crypto.randomUUID(),
        requestId: request.id,
        iteration,
        name: request.name || request.url,
        method: request.method,
        url: response?.requestedUrl || request.url,
        status: passed ? 'passed' : 'failed',
        statusCode: response?.statusCode ?? 0,
        durationMs: response?.durationMs ?? 0,
        message,
        request: response?.request,
        response: response ?? new domain.Response({ error: message }),
        testResults: response?.testResults ?? [],
        startedAt: new Date().toISOString(),
        finishedAt: new Date().toISOString()
      })
    ]
  }

  async function persistRunnerResult() {
    if (!runnerResult.value) return
    try {
      const next = await options.saveRunnerResult(runnerResult.value)
      options.setState(next)
    } catch (error) {
      options.statusMessage.value = formatError(error)
    }
  }

  function updateRunnerQueueItem(id: string, patch: Partial<RunnerQueueItem>) {
    runnerQueue.value = patchRunnerQueueItem(runnerQueue.value, id, patch)
  }

  function runnerStatusMessage() {
    return `${options.labels.value.runner}: ${runnerResult.value?.passed ?? 0} ${options.labels.value.passed}, ${runnerResult.value?.failed ?? 0} ${options.labels.value.failed}`
  }

  function runnerReportText() {
    const result = runnerResult.value
    const lines = [
      `# ${options.labels.value.runner}`,
      '',
      `${options.labels.value.passed}: ${result?.passed ?? 0}`,
      `${options.labels.value.failed}: ${result?.failed ?? 0}`,
      `Duration: ${result?.durationMs ?? 0} ms`,
      '',
      '| # | Iteration | Method | Request | URL | Status | Code | Duration | Message |',
      '| --- | --- | --- | --- | --- | --- | --- | --- | --- |'
    ]
    runnerQueue.value.forEach((item, index) => {
      lines.push(`| ${index + 1} | ${item.iteration} | ${item.method} | ${escapeCell(item.name)} | ${escapeCell(item.url)} | ${item.status} | ${item.statusCode ?? '-'} | ${item.durationMs != null ? `${item.durationMs} ms` : '-'} | ${escapeCell(item.message || '-')} |`)
    })
    const details = result?.details ?? []
    if (details.length) {
      lines.push('', '## Details')
      details.forEach((detail, index) => {
        lines.push(
          '',
          `### ${index + 1}. ${detail.method} ${detail.name}`,
          '',
          `- Iteration: ${detail.iteration}`,
          `- URL: ${detail.url}`,
          `- Status: ${detail.status}`,
          `- Code: ${detail.statusCode || '-'}`,
          `- Duration: ${detail.durationMs} ms`,
          `- Message: ${detail.message || '-'}`,
          '',
          '#### Request',
          '',
          `${detail.request?.method || detail.method} ${detail.request?.url || detail.url}`
        )
        if (detail.request?.headers?.length) {
          lines.push('', '| Header | Value |', '| --- | --- |')
          detail.request.headers.forEach((header) => {
            lines.push(`| ${escapeCell(header.key)} | ${escapeCell(header.value)} |`)
          })
        }
        if (detail.request?.body?.text) {
          lines.push('', '```text', detail.request.body.text, '```')
        }
        lines.push('', '#### Response', '', detail.response?.status || detail.response?.error || '-')
        if (detail.response?.body) {
          lines.push('', '```text', detail.response.body, '```')
        }
        if (detail.testResults?.length) {
          lines.push('', '#### Tests', '', '| Test | Result | Message |', '| --- | --- | --- |')
          detail.testResults.forEach((test) => {
            lines.push(`| ${escapeCell(test.name)} | ${test.passed ? 'passed' : 'failed'} | ${escapeCell(test.message || '-')} |`)
          })
        }
      })
    }
    return lines.join('\n')
  }

  function delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }

  function escapeCell(value: string) {
    return String(value).replaceAll('|', '\\|').replace(/\r?\n/g, ' ')
  }

  return {
    runnerBusy,
    runnerResult,
    runnerScope,
    runnerIterations,
    runnerRequestId,
    runnerQueue,
    runnerFailurePolicy,
    runnerStopRequested,
    runnerRetryCount,
    runnerDelayMs,
    ensureRunnerRequest,
    setRunnerRequest,
    selectRunnerRequest,
    setRunnerScope,
    setRunnerIterations,
    setRunnerFailurePolicy,
    setRunnerRetryCount,
    setRunnerDelayMs,
    stopRunner,
    resetRunnerOutput,
    runActiveCollection,
    runRunnerRequest,
    runnerReportText
  }
}
