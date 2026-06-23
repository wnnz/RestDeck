# RestDeck

RestDeck 是一个本地优先的桌面 API 调试工具，专注于常用 HTTP 请求、环境变量、历史记录和集合运行工作流。

## 界面截图

### 请求工作台

![请求工作台](docs/images/restdeck-requests-light.png)

### 暗色模式

![暗色模式请求工作台](docs/images/restdeck-requests-dark.png)

### 环境变量

![环境变量](docs/images/restdeck-environments-light.png)

### 集合 Runner

![集合 Runner](docs/images/restdeck-runner-light.png)

### 设置

![暗色模式设置](docs/images/restdeck-settings-dark.png)

## 功能

### 请求调试

- 支持创建、复制、置顶、删除和导出请求。
- 支持 GET、POST、PUT、PATCH、DELETE 等常用 HTTP Method。
- 支持 URL、查询参数、请求头、Body、认证、预请求脚本、测试脚本和超时时间配置。
- 支持 Params、Headers 等表格化编辑，并自动保存修改。
- 支持 JSON Body 编辑、美化和语法高亮。
- 支持 form 请求体，包含文本字段和本地文件上传字段。
- 支持从浏览器 Copy as Fetch / Copy as cURL 导入请求。
- 支持生成常用客户端代码，包括 cURL、Fetch、Node.js、Python、Go、C#、Java 等格式。

### 响应查看

- 展示响应状态码、耗时、响应大小和内容类型。
- 支持查看响应 Body、响应头、Cookies 和测试结果。
- 支持响应正文的原始视图、格式化视图和 JSON 高亮。
- 历史记录展示实际请求地址、状态和耗时，可回到对应请求查看。

### 集合与 Runner

- 支持集合创建、重命名、删除、导入和导出。
- 支持 Postman Collection JSON 导入与导出。
- 支持按集合或单个请求运行。
- Runner 可展示等待中、运行中、通过、失败等请求状态。
- Runner 支持迭代次数、当前环境和运行结果统计。

### 环境与变量

- 支持全局变量和环境变量。
- 支持新增、重命名、删除环境。
- 环境变量支持静态值、时间戳和从请求响应 JSONPath 读取值。
- 响应变量支持读取最新历史、每次读取前请求、超时后重新请求。
- 输入框支持 `{{变量}}` 提示和插入。
- 支持动态变量，如 `$guid`、`$timestamp`、`$isoTimestamp`、`$randomInt`、`$randomBoolean`、`$randomEmail` 等。
- 环境和变量修改自动保存。

### 认证、代理与脚本

- 支持 No Auth、API Key、Bearer Token、Basic、Digest、OAuth 1、OAuth 2 Bearer Token 等认证配置。
- 支持默认代理、请求独立代理、禁用代理和代理排除规则。
- 支持 HTTP、WebSocket、SSE 使用代理配置。
- 支持常用 Postman `pm` 脚本子集，包括变量读取替换、请求对象、响应对象、测试断言等。

### 实时与其他能力

- 支持 WebSocket 单消息调试。
- 支持 SSE 事件流采集。
- 支持浅色和暗色主题切换。
- 支持中文和英文界面。
- 数据保存到程序同目录的 `Data` 文件夹。
- 敏感字段使用本地加密封装保存。
- 支持 Windows 自定义标题栏、窗口拖动、最小化、最大化和关闭。
