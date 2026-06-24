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
      const result = await options.sendRequest(requestToSend, options.activeEnvironment.value?.id ?? '', options.globalsDraft.value)
      const summary = summarizeRunnerResponse(result, options.labels.value.tests)
      updateRunnerQueueItem(queueId, {
        status: summary.passed ? 'passed' : 'failed',
        url: result.requestedUrl || requestToSend.url,
        statusCode: result.statusCode,
        durationMs: result.durationMs,
        message: summary.message
      })
      addRunnerResultItem(requestToSend, summary.passed, summary.message)
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
      addRunnerResultItem(requestToSend, false, message)
      return false
    }
    return runnerQueue.value.find((item) => item.id === queueId)?.status === 'passed'
  }

  function skipPendingRunnerItems(message: string) {
    runnerQueue.value = markPendingRunnerItems(runnerQueue.value, {
      status: 'skipped',
      message
    })
  }

  function addRunnerResultItem(request: domain.Request, passed: boolean, message: string) {
    if (!runnerResult.value) return
    runnerResult.value.passed += passed ? 1 : 0
    runnerResult.value.failed += passed ? 0 : 1
    runnerResult.value.items = [
      ...(runnerResult.value.items ?? []),
      new domain.TestResult({ name: request.name || request.url, passed, message })
    ]
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
    return lines.join('\n')
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
    ensureRunnerRequest,
    setRunnerRequest,
    selectRunnerRequest,
    setRunnerScope,
    setRunnerIterations,
    setRunnerFailurePolicy,
    stopRunner,
    resetRunnerOutput,
    runActiveCollection,
    runRunnerRequest,
    runnerReportText
  }
}
