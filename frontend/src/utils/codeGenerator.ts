import { domain } from '../../wailsjs/go/models'
import { cloneRequest, normalizeRequest, syncFormBody } from './requestModel'

export type CodeFormat =
  | 'curl-linux'
  | 'curl-powershell'
  | 'javascript-fetch'
  | 'node-fetch'
  | 'python-requests'
  | 'go-http'
  | 'csharp-httpclient'
  | 'java-httpclient'

export type CodeFormatOption = {
  value: CodeFormat
  label: string
}

export const codeFormatOptions: CodeFormatOption[] = [
  { value: 'curl-linux', label: 'cURL Linux/macOS' },
  { value: 'curl-powershell', label: 'cURL PowerShell' },
  { value: 'javascript-fetch', label: 'JavaScript fetch' },
  { value: 'node-fetch', label: 'Node.js fetch' },
  { value: 'python-requests', label: 'Python requests' },
  { value: 'go-http', label: 'Go net/http' },
  { value: 'csharp-httpclient', label: 'C# HttpClient' },
  { value: 'java-httpclient', label: 'Java HttpClient' }
]

type BodyMode = 'none' | 'json' | 'raw' | 'form' | 'urlencoded'

type Header = {
  key: string
  value: string
}

type FormPart = {
  key: string
  type: 'text' | 'file'
  value: string
}

type PreparedRequest = {
  method: string
  url: string
  headers: Header[]
  bodyMode: BodyMode
  body: string
  formParts: FormPart[]
  urlencoded: Header[]
  auth: domain.AuthConfig
  oauth1: boolean
}

export function generateRequestCode(request: domain.Request, format: CodeFormat) {
  const prepared = prepareRequest(request)
  switch (format) {
    case 'curl-powershell':
      return generateCurlPowerShell(prepared)
    case 'javascript-fetch':
      return generateJavaScriptFetch(prepared, false)
    case 'node-fetch':
      return generateJavaScriptFetch(prepared, true)
    case 'python-requests':
      return generatePythonRequests(prepared)
    case 'go-http':
      return generateGoHTTP(prepared)
    case 'csharp-httpclient':
      return generateCSharpHttpClient(prepared)
    case 'java-httpclient':
      return generateJavaHttpClient(prepared)
    case 'curl-linux':
    default:
      return generateCurlLinux(prepared)
  }
}

function prepareRequest(input: domain.Request): PreparedRequest {
  const request = normalizeRequest(cloneRequest(input))
  syncFormBody(request)
  const method = (request.method || 'GET').trim().toUpperCase()
  let url = request.url || ''
  const headers: Header[] = []
  for (const param of request.params ?? []) {
    if (param.enabled && param.key) {
      url = appendQuery(url, param.key, param.value ?? '')
    }
  }
  for (const header of request.headers ?? []) {
    if (header.enabled && header.key) {
      setHeader(headers, header.key, header.value ?? '')
    }
  }

  const bodyMode = normalizeBodyMode(request.bodyMode)
  const body = request.body ?? ''
  const formParts = bodyMode === 'form'
    ? (request.formItems ?? [])
      .filter((item) => item.enabled && item.key)
      .map((item) => ({
        key: item.key,
        type: item.type === 'file' ? 'file' as const : 'text' as const,
        value: item.type === 'file' ? (item.filePath ?? '') : (item.value ?? '')
      }))
    : []
  const urlencoded = bodyMode === 'urlencoded' ? parseBodyRows(body) : []

  if (bodyMode === 'json' && !hasHeader(headers, 'Content-Type')) {
    setHeader(headers, 'Content-Type', 'application/json')
  }
  if (bodyMode === 'urlencoded' && !hasHeader(headers, 'Content-Type')) {
    setHeader(headers, 'Content-Type', 'application/x-www-form-urlencoded')
  }

  const auth = request.auth ?? new domain.AuthConfig({ type: 'none', values: {} })
  const authValues = auth.values ?? {}
  let oauth1 = false
  switch (auth.type) {
    case 'apiKey':
      if (authValues.key) {
        if (authValues.in === 'query') {
          url = appendQuery(url, authValues.key, authValues.value ?? '')
        } else {
          setHeader(headers, authValues.key, authValues.value ?? '')
        }
      }
      break
    case 'bearer':
      if (authValues.token) setHeader(headers, 'Authorization', `Bearer ${authValues.token}`)
      break
    case 'oauth2':
      if (authValues.accessToken) setHeader(headers, 'Authorization', `Bearer ${authValues.accessToken}`)
      break
    case 'digest':
      if (authValues.username || authValues.password) {
        setHeader(headers, 'Authorization', `Digest username="${escapeHeader(authValues.username ?? '')}", password="${escapeHeader(authValues.password ?? '')}"`)
      }
      break
    case 'oauth1':
      oauth1 = true
      break
  }

  return { method, url, headers, bodyMode, body, formParts, urlencoded, auth, oauth1 }
}

function normalizeBodyMode(value: string): BodyMode {
  if (value === 'json' || value === 'raw' || value === 'form' || value === 'urlencoded') return value
  return 'none'
}

function appendQuery(url: string, key: string, value: string) {
  const separator = url.includes('?')
    ? (url.endsWith('?') || url.endsWith('&') ? '' : '&')
    : '?'
  return `${url}${separator}${key}=${value}`
}

function parseBodyRows(raw: string) {
  return raw
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const index = line.indexOf('=')
      return {
        key: index >= 0 ? line.slice(0, index).trim() : line,
        value: index >= 0 ? line.slice(index + 1).trim() : ''
      }
    })
    .filter((row) => row.key)
}

function hasHeader(headers: Header[], key: string) {
  return headers.some((header) => header.key.toLowerCase() === key.toLowerCase())
}

function setHeader(headers: Header[], key: string, value: string) {
  const existing = headers.find((header) => header.key.toLowerCase() === key.toLowerCase())
  if (existing) {
    existing.key = key
    existing.value = value
    return
  }
  headers.push({ key, value })
}

function escapeHeader(value: string) {
  return value.replace(/\\/g, '\\\\').replace(/"/g, '\\"')
}

function shQuote(value: string) {
  return `'${value.replace(/'/g, `'\"'\"'`)}'`
}

function psQuote(value: string) {
  return `'${value.replace(/'/g, "''")}'`
}

function jsString(value: string) {
  return JSON.stringify(value)
}

function pyString(value: string) {
  return JSON.stringify(value)
}

function goString(value: string) {
  return JSON.stringify(value)
}

function csString(value: string) {
  return JSON.stringify(value)
}

function javaString(value: string) {
  return JSON.stringify(value)
}

function baseName(path: string) {
  return path.split(/[\\/]/).filter(Boolean).pop() || 'file'
}

function basicAuthPair(request: PreparedRequest) {
  if (request.auth?.type !== 'basic') return null
  const values = request.auth.values ?? {}
  if (!values.username && !values.password) return null
  return `${values.username ?? ''}:${values.password ?? ''}`
}

function oauth1Note(request: PreparedRequest, prefix = '#') {
  return request.oauth1 ? `${prefix} OAuth 1.0 signatures depend on nonce/timestamp and are not expanded by this snippet.\n` : ''
}

function appendBodyCurl(lines: string[], request: PreparedRequest, quote: (value: string) => string) {
  if (request.bodyMode === 'form') {
    for (const part of request.formParts) {
      lines.push(`  -F ${quote(`${part.key}=${part.type === 'file' ? `@${part.value}` : part.value}`)}`)
    }
    return
  }
  if (request.bodyMode === 'urlencoded') {
    for (const row of request.urlencoded) {
      lines.push(`  --data-urlencode ${quote(`${row.key}=${row.value}`)}`)
    }
    return
  }
  if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    lines.push(`  --data-raw ${quote(request.body)}`)
  }
}

function generateCurlLinux(request: PreparedRequest) {
  const lines = [`curl -X ${request.method} ${shQuote(request.url)}`]
  for (const header of request.headers) {
    lines.push(`  -H ${shQuote(`${header.key}: ${header.value}`)}`)
  }
  const basic = basicAuthPair(request)
  if (basic) lines.push(`  -u ${shQuote(basic)}`)
  appendBodyCurl(lines, request, shQuote)
  return `${oauth1Note(request)}${lines.join(' \\\n')}`
}

function generateCurlPowerShell(request: PreparedRequest) {
  const lines = [`curl.exe -X ${request.method} ${psQuote(request.url)}`]
  for (const header of request.headers) {
    lines.push(`  -H ${psQuote(`${header.key}: ${header.value}`)}`)
  }
  const basic = basicAuthPair(request)
  if (basic) lines.push(`  -u ${psQuote(basic)}`)
  appendBodyCurl(lines, request, psQuote)
  return `${oauth1Note(request)}${lines.join(' `\n')}`
}

function jsHeadersExpression(request: PreparedRequest, node: boolean) {
  const lines = request.headers.map((header) => `  ${jsString(header.key)}: ${jsString(header.value)}`)
  if (!lines.length && !basicAuthPair(request)) return 'const headers = new Headers();'
  const body = lines.length ? `{\n${lines.join(',\n')}\n}` : '{}'
  const out = [`const headers = new Headers(${body});`]
  const basic = basicAuthPair(request)
  if (basic) {
    const expression = node
      ? `Buffer.from(${jsString(basic)}).toString('base64')`
      : `btoa(${jsString(basic)})`
    out.push(`headers.set('Authorization', 'Basic ' + ${expression});`)
  }
  return out.join('\n')
}

function jsBodyLines(request: PreparedRequest, node: boolean) {
  if (request.bodyMode === 'form') {
    const lines = node ? [`import { readFileSync } from 'node:fs';`, '', 'const formData = new FormData();'] : ['const formData = new FormData();']
    for (const part of request.formParts) {
      if (part.type === 'file') {
        if (node) {
          lines.push(`formData.append(${jsString(part.key)}, new Blob([readFileSync(${jsString(part.value)})]), ${jsString(baseName(part.value))});`)
        } else {
          lines.push(`formData.append(${jsString(part.key)}, fileInput.files[0]); // ${part.value}`)
        }
      } else {
        lines.push(`formData.append(${jsString(part.key)}, ${jsString(part.value)});`)
      }
    }
    return { prelude: lines.join('\n'), body: 'formData' }
  }
  if (request.bodyMode === 'urlencoded') {
    const entries = request.urlencoded.map((row) => `  [${jsString(row.key)}, ${jsString(row.value)}]`)
    return {
      prelude: `const body = new URLSearchParams([\n${entries.join(',\n')}\n]);`,
      body: 'body'
    }
  }
  if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    return { prelude: `const body = ${jsString(request.body)};`, body: 'body' }
  }
  return { prelude: '', body: '' }
}

function generateJavaScriptFetch(request: PreparedRequest, node: boolean) {
  const body = jsBodyLines(request, node)
  const lines = [
    oauth1Note(request, '//').trimEnd(),
    body.prelude,
    jsHeadersExpression(request, node),
    '',
    `const response = await fetch(${jsString(request.url)}, {`,
    `  method: ${jsString(request.method)},`,
    '  headers,'
  ].filter(Boolean)
  if (body.body) lines.push(`  body: ${body.body},`)
  lines.push('});', '', 'const data = await response.text();', 'console.log(data);')
  return lines.join('\n')
}

function pyDict(items: Header[], indent = '    ') {
  if (!items.length) return '{}'
  return `{\n${items.map((item) => `${indent}${pyString(item.key)}: ${pyString(item.value)}`).join(',\n')}\n}`
}

function generatePythonRequests(request: PreparedRequest) {
  const lines = [oauth1Note(request).trimEnd(), 'import requests', '', `url = ${pyString(request.url)}`, `headers = ${pyDict(request.headers)}`].filter(Boolean)
  const args = [`${pyString(request.method)}`, 'url', 'headers=headers']
  const basic = basicAuthPair(request)
  if (basic) {
    const index = basic.indexOf(':')
    args.push(`auth=(${pyString(basic.slice(0, index))}, ${pyString(basic.slice(index + 1))})`)
  }
  if (request.bodyMode === 'form') {
    const textParts = request.formParts.filter((part) => part.type === 'text').map((part) => ({ key: part.key, value: part.value }))
    const fileParts = request.formParts.filter((part) => part.type === 'file')
    lines.push(`data = ${pyDict(textParts)}`)
    lines.push(`files = ${fileParts.length ? `{\n${fileParts.map((part) => `    ${pyString(part.key)}: (${pyString(baseName(part.value))}, open(${pyString(part.value)}, "rb"))`).join(',\n')}\n}` : '{}'}`)
    args.push('data=data', 'files=files')
  } else if (request.bodyMode === 'urlencoded') {
    lines.push(`data = ${pyDict(request.urlencoded)}`)
    args.push('data=data')
  } else if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    args.push(`data=${pyString(request.body)}`)
  }
  lines.push('', `response = requests.request(${args.join(', ')})`, 'print(response.text)')
  return lines.join('\n')
}

function goImports(imports: Set<string>) {
  return `import (\n${Array.from(imports).sort().map((item) => `\t${goString(item)}`).join('\n')}\n)`
}

function generateGoHTTP(request: PreparedRequest) {
  const imports = new Set(['fmt', 'io', 'net/http'])
  const lines = [oauth1Note(request, '//').trimEnd(), 'package main', ''].filter(Boolean)
  let bodyReader = 'nil'
  const beforeRequest: string[] = []
  const afterRequest: string[] = []
  if (request.bodyMode === 'form') {
    imports.add('bytes')
    imports.add('mime/multipart')
    imports.add('os')
    imports.add('path/filepath')
    beforeRequest.push('var requestBody bytes.Buffer', 'writer := multipart.NewWriter(&requestBody)')
    for (const part of request.formParts) {
      if (part.type === 'file') {
        beforeRequest.push(
          `file, err := os.Open(${goString(part.value)})`,
          'if err != nil { panic(err) }',
          'defer file.Close()',
          `filePart, err := writer.CreateFormFile(${goString(part.key)}, filepath.Base(${goString(part.value)}))`,
          'if err != nil { panic(err) }',
          '_, _ = io.Copy(filePart, file)'
        )
      } else {
        beforeRequest.push(`_ = writer.WriteField(${goString(part.key)}, ${goString(part.value)})`)
      }
    }
    beforeRequest.push('writer.Close()')
    bodyReader = '&requestBody'
    afterRequest.push('request.Header.Set("Content-Type", writer.FormDataContentType())')
  } else if (request.bodyMode === 'urlencoded') {
    imports.add('net/url')
    imports.add('strings')
    beforeRequest.push('form := url.Values{}')
    for (const row of request.urlencoded) {
      beforeRequest.push(`form.Set(${goString(row.key)}, ${goString(row.value)})`)
    }
    beforeRequest.push('body := strings.NewReader(form.Encode())')
    bodyReader = 'body'
  } else if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    imports.add('strings')
    beforeRequest.push(`body := strings.NewReader(${goString(request.body)})`)
    bodyReader = 'body'
  }
  lines.push(goImports(imports), '', 'func main() {')
  lines.push(...beforeRequest.map((line) => `\t${line}`))
  lines.push(`\trequest, err := http.NewRequest(${goString(request.method)}, ${goString(request.url)}, ${bodyReader})`)
  lines.push('\tif err != nil { panic(err) }')
  for (const header of request.headers) {
    lines.push(`\trequest.Header.Set(${goString(header.key)}, ${goString(header.value)})`)
  }
  lines.push(...afterRequest.map((line) => `\t${line}`))
  const basic = basicAuthPair(request)
  if (basic) {
    const index = basic.indexOf(':')
    lines.push(`\trequest.SetBasicAuth(${goString(basic.slice(0, index))}, ${goString(basic.slice(index + 1))})`)
  }
  lines.push(
    '\tresponse, err := http.DefaultClient.Do(request)',
    '\tif err != nil { panic(err) }',
    '\tdefer response.Body.Close()',
    '\tbodyBytes, _ := io.ReadAll(response.Body)',
    '\tfmt.Println(string(bodyBytes))',
    '}'
  )
  return lines.join('\n')
}

function generateCSharpHttpClient(request: PreparedRequest) {
  const lines = [
    oauth1Note(request, '//').trimEnd(),
    'using System;',
    'using System.Collections.Generic;',
    'using System.IO;',
    'using System.Net.Http;',
    'using System.Net.Http.Headers;',
    'using System.Text;',
    '',
    'using var client = new HttpClient();',
    `using var request = new HttpRequestMessage(new HttpMethod(${csString(request.method)}), ${csString(request.url)});`
  ].filter(Boolean)
  for (const header of request.headers) {
    if (header.key.toLowerCase() !== 'content-type') {
      lines.push(`request.Headers.TryAddWithoutValidation(${csString(header.key)}, ${csString(header.value)});`)
    }
  }
  const basic = basicAuthPair(request)
  if (basic) {
    lines.push(`request.Headers.Authorization = new AuthenticationHeaderValue("Basic", Convert.ToBase64String(Encoding.UTF8.GetBytes(${csString(basic)})));`)
  }
  if (request.bodyMode === 'form') {
    lines.push('using var content = new MultipartFormDataContent();')
    for (const part of request.formParts) {
      if (part.type === 'file') {
        lines.push(`content.Add(new StreamContent(File.OpenRead(${csString(part.value)})), ${csString(part.key)}, ${csString(baseName(part.value))});`)
      } else {
        lines.push(`content.Add(new StringContent(${csString(part.value)}), ${csString(part.key)});`)
      }
    }
    lines.push('request.Content = content;')
  } else if (request.bodyMode === 'urlencoded') {
    lines.push('request.Content = new FormUrlEncodedContent(new Dictionary<string, string>')
    lines.push('{')
    for (const row of request.urlencoded) {
      lines.push(`    [${csString(row.key)}] = ${csString(row.value)},`)
    }
    lines.push('});')
  } else if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    const contentType = request.headers.find((header) => header.key.toLowerCase() === 'content-type')?.value || 'text/plain'
    lines.push(`request.Content = new StringContent(${csString(request.body)}, Encoding.UTF8, ${csString(contentType)});`)
  }
  lines.push('', 'using var response = await client.SendAsync(request);', 'Console.WriteLine(await response.Content.ReadAsStringAsync());')
  return lines.join('\n')
}

function generateJavaHttpClient(request: PreparedRequest) {
  const lines = [
    oauth1Note(request, '//').trimEnd(),
    'import java.net.URI;',
    'import java.net.http.HttpClient;',
    'import java.net.http.HttpRequest;',
    'import java.net.http.HttpResponse;',
    'import java.nio.charset.StandardCharsets;',
    'import java.util.Base64;',
    '',
    'HttpClient client = HttpClient.newHttpClient();',
    `HttpRequest.Builder builder = HttpRequest.newBuilder().uri(URI.create(${javaString(request.url)}));`
  ].filter(Boolean)
  for (const header of request.headers) {
    lines.push(`builder.header(${javaString(header.key)}, ${javaString(header.value)});`)
  }
  const basic = basicAuthPair(request)
  if (basic) {
    lines.push(`builder.header("Authorization", "Basic " + Base64.getEncoder().encodeToString(${javaString(basic)}.getBytes(StandardCharsets.UTF_8)));`)
  }
  if (request.bodyMode === 'urlencoded') {
    const body = request.urlencoded.map((row) => `${row.key}=${row.value}`).join('&')
    lines.push(`builder.method(${javaString(request.method)}, HttpRequest.BodyPublishers.ofString(${javaString(body)}));`)
  } else if (request.bodyMode === 'form') {
    lines.push('// Multipart file upload in Java requires building a multipart body with boundaries or using a helper library.')
    lines.push(`builder.method(${javaString(request.method)}, HttpRequest.BodyPublishers.noBody());`)
  } else if ((request.bodyMode === 'json' || request.bodyMode === 'raw') && request.body) {
    lines.push(`builder.method(${javaString(request.method)}, HttpRequest.BodyPublishers.ofString(${javaString(request.body)}));`)
  } else {
    lines.push(`builder.method(${javaString(request.method)}, HttpRequest.BodyPublishers.noBody());`)
  }
  lines.push('', 'HttpResponse<String> response = client.send(builder.build(), HttpResponse.BodyHandlers.ofString());', 'System.out.println(response.body());')
  return lines.join('\n')
}
