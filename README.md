# RestDeck

RestDeck 是一个本地优先、开源、无账号登录的桌面 API 测试工具，目标是覆盖 Postman 中常用的本地调试工作流。它不依赖云同步、团队空间或账号体系；应用内只展示已经实现的功能，未实现的能力不会出现在导航、按钮或占位面板里。

## 技术栈

- Go + Wails v2
- Vue 3 + Vite
- Tailwind CSS v4
- PrimeVue Volt 风格的代码自有 UI
- SQLite 保存本地工作区数据
- goja 支持常用 Postman `pm` 脚本子集

## 当前已实现

- 集合、请求、环境、全局变量、历史记录、Runner、本地设置页面。
- 集合下拉选择、集合改名、集合删除，以及新建请求菜单。
- HTTP/REST 请求编辑：Method、URL、Params、Headers、Body、Auth、Pre-request、Tests、Timeout。
- 常见认证：No Auth、API Key、Bearer Token、Basic、Digest 配置占位、OAuth 1 签名、OAuth 2 Bearer Token。
- 响应查看：Body、Headers、Cookies、Test Results、状态码、耗时、大小。
- JSON 响应美化与基础语法着色。
- 本地 SQLite 持久化，数据位于系统用户配置目录。
- 敏感环境变量和认证字段使用本地加密封装保存。
- Postman Collection JSON 常用字段导入/导出。
- 浏览器 Copy as Fetch 导入请求。
- 浏览器 Copy as cURL 常见格式导入请求。
- 动态变量：`{{$guid}}`、`{{$timestamp}}`、`{{$isoTimestamp}}`、`{{$randomInt}}`、`{{$randomBoolean}}`、`{{$randomEmail}}`。
- 常用 `pm` 脚本子集：`pm.variables.get/set/replaceIn`、`pm.request`、`pm.response`、`pm.test`、基础 `expect(...)` 断言。
- WebSocket 单消息调试和 SSE 事件流采集。
- Windows 下自定义标题栏，支持最小化、最大化/还原、关闭、拖动和双击最大化/还原。

## 暂未实现

这些功能在真正实现前不会显示在应用 UI 中：

- gRPC
- Mock Servers
- Monitors
- Flows
- 团队/云工作区
- AI 功能
- 完整 Postman Sandbox 兼容
- 完整 Digest challenge-response 流程
- OpenAPI 导入/导出
- 请求文档侧边栏

当前功能状态见 [docs/FEATURE_MATRIX.md](docs/FEATURE_MATRIX.md)。

## 开发

如本机尚未安装 Wails v2：

```powershell
$env:GOPROXY='https://goproxy.cn,direct'
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

安装前端依赖：

```powershell
cd frontend
npm install
```

运行检查：

```powershell
go test ./...
cd frontend
npm run build
```

构建桌面应用：

```powershell
$env:Path="$env:USERPROFILE\go\bin;$env:Path"
wails build -clean
```

开发模式运行：

```powershell
$env:Path="$env:USERPROFILE\go\bin;$env:Path"
wails dev
```

## 产品方向

RestDeck 会优先打磨高频、本地、无账号的 API 调试体验，再逐步扩展到更完整的非账号功能集。核心原则是：界面紧凑、专业、诚实；功能没完成，就不在应用里假装存在。
