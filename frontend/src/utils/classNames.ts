import { twMerge } from 'tailwind-merge'
import type { ClassValue } from 'vue'

function flattenClass(value: ClassValue | false | null | undefined): string {
  if (!value) return ''
  if (typeof value === 'string') return value
  if (Array.isArray(value)) return value.map((item) => flattenClass(item)).filter(Boolean).join(' ')
  if (typeof value === 'object') {
    return Object.entries(value)
      .filter(([, enabled]) => enabled)
      .map(([name]) => name)
      .join(' ')
  }
  return ''
}

export function cn(...classes: Array<ClassValue | false | null | undefined>) {
  return twMerge(classes.map((item) => flattenClass(item)).filter(Boolean).join(' '))
}
