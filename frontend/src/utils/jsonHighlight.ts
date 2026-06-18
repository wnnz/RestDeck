import type { JsonToken, JsonTokenType } from '../types'

export function tokenizeJSON(raw: string): JsonToken[] {
  const tokens: JsonToken[] = []
  const pattern = /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(?=\s*:))|("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|\b(true|false)\b|\bnull\b|([{}[\],:])/g
  let lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = pattern.exec(raw)) !== null) {
    if (match.index > lastIndex) {
      tokens.push({ type: 'plain', text: raw.slice(lastIndex, match.index) })
    }
    const text = match[0]
    let type: JsonTokenType = 'plain'
    if (match[1]) type = 'key'
    else if (match[2]) type = 'string'
    else if (match[3]) type = 'number'
    else if (match[4]) type = 'boolean'
    else if (text === 'null') type = 'null'
    else if (match[5]) type = 'punctuation'
    tokens.push({ type, text })
    lastIndex = pattern.lastIndex
  }
  if (lastIndex < raw.length) {
    tokens.push({ type: 'plain', text: raw.slice(lastIndex) })
  }
  return tokens
}
