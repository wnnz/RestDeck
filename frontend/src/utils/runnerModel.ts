import { domain } from '../../wailsjs/go/models'
import type { RunnerQueueItem } from '../types'

export function newRunnerResult(input: { collectionId: string; environmentId: string; name: string; iterations: number }) {
  return new domain.RunnerResult({
    id: crypto.randomUUID(),
    collectionId: input.collectionId,
    environmentId: input.environmentId,
    name: input.name,
    iterations: input.iterations,
    passed: 0,
    failed: 0,
    durationMs: 0,
    items: [],
    details: [],
    createdAt: new Date().toISOString()
  })
}

export function buildRunnerQueue(requests: domain.Request[], iterations: number) {
  return Array.from({ length: iterations }).flatMap((_, iterationIndex) => requests.map((request) => ({
    id: `${request.id}-${iterationIndex + 1}`,
    requestId: request.id,
    iteration: iterationIndex + 1,
    method: request.method,
    name: request.name || request.url,
    url: request.url,
      status: 'waiting' as const
  })))
}

export function markPendingRunnerItems(items: RunnerQueueItem[], patch: Partial<RunnerQueueItem>) {
  return items.map((item) => item.status === 'waiting' ? { ...item, ...patch } : item)
}

export function summarizeRunnerResponse(result: domain.Response, testsLabel: string) {
  if (result.error) {
    return { passed: false, message: result.error }
  }
  const tests = result.testResults ?? []
  if (tests.length) {
    const failedTest = tests.find((item) => !item.passed)
    if (failedTest) {
      return {
        passed: false,
        message: failedTest.message ? `${failedTest.name}: ${failedTest.message}` : failedTest.name
      }
    }
    return { passed: true, message: `${tests.length} ${testsLabel}` }
  }
  return {
    passed: result.statusCode >= 200 && result.statusCode < 400,
    message: result.status || `${result.statusCode || '-'}`
  }
}

export function patchRunnerQueueItem(items: RunnerQueueItem[], id: string, patch: Partial<RunnerQueueItem>) {
  let updated = false
  const next = items.map((item) => {
    if (item.id !== id) return item
    updated = true
    return { ...item, ...patch }
  })
  if (updated) return next
  return [
    ...next,
    {
      id,
      requestId: id,
      iteration: 1,
      method: '-',
      name: id,
      url: '',
      status: 'failed' as const,
      ...patch
    }
  ]
}
