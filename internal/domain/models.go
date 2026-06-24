package domain

import "time"

type KeyValue struct {
	ID                  string `json:"id"`
	Enabled             bool   `json:"enabled"`
	Key                 string `json:"key"`
	Value               string `json:"value"`
	Description         string `json:"description"`
	Secret              bool   `json:"secret"`
	ValueType           string `json:"valueType"`
	TimestampFormat     string `json:"timestampFormat"`
	SourceRequestID     string `json:"sourceRequestId"`
	JSONPath            string `json:"jsonPath"`
	ResponseStrategy    string `json:"responseStrategy"`
	RefreshAfterSeconds int    `json:"refreshAfterSeconds"`
	FallbackValue       string `json:"fallbackValue"`
}

type FormItem struct {
	ID          string `json:"id"`
	Enabled     bool   `json:"enabled"`
	Key         string `json:"key"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	FilePath    string `json:"filePath"`
	Description string `json:"description"`
}

type BodyMode string

const (
	BodyModeNone       BodyMode = "none"
	BodyModeRaw        BodyMode = "raw"
	BodyModeJSON       BodyMode = "json"
	BodyModeForm       BodyMode = "form"
	BodyModeURLEncoded BodyMode = "urlencoded"
)

type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeAPIKey AuthType = "apiKey"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeDigest AuthType = "digest"
	AuthTypeOAuth1 AuthType = "oauth1"
	AuthTypeOAuth2 AuthType = "oauth2"
)

type AuthConfig struct {
	Type   AuthType          `json:"type"`
	Values map[string]string `json:"values"`
}

type ProxyConfig struct {
	Mode    string `json:"mode"`
	URL     string `json:"url"`
	NoProxy string `json:"noProxy"`
}

type Request struct {
	ID           string      `json:"id"`
	CollectionID string      `json:"collectionId"`
	ParentID     string      `json:"parentId"`
	Name         string      `json:"name"`
	Method       string      `json:"method"`
	URL          string      `json:"url"`
	Params       []KeyValue  `json:"params"`
	Headers      []KeyValue  `json:"headers"`
	BodyMode     BodyMode    `json:"bodyMode"`
	Body         string      `json:"body"`
	FormItems    []FormItem  `json:"formItems"`
	Auth         AuthConfig  `json:"auth"`
	Proxy        ProxyConfig `json:"proxy"`
	PreScript    string      `json:"preScript"`
	TestScript   string      `json:"testScript"`
	TimeoutMs    int         `json:"timeoutMs"`
	SortOrder    int         `json:"sortOrder"`
	UpdatedAt    time.Time   `json:"updatedAt" ts_type:"string"`
}

type Folder struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collectionId"`
	ParentID     string    `json:"parentId"`
	Name         string    `json:"name"`
	SortOrder    int       `json:"sortOrder"`
	UpdatedAt    time.Time `json:"updatedAt" ts_type:"string"`
}

type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Folders     []Folder  `json:"folders"`
	Requests    []Request `json:"requests"`
	CreatedAt   time.Time `json:"createdAt" ts_type:"string"`
	UpdatedAt   time.Time `json:"updatedAt" ts_type:"string"`
}

type Environment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Variables []KeyValue `json:"variables"`
	IsActive  bool       `json:"isActive"`
	UpdatedAt time.Time  `json:"updatedAt" ts_type:"string"`
}

type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires" ts_type:"string"`
	HTTPOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}

type PreparedBody struct {
	Mode        BodyMode   `json:"mode"`
	ContentType string     `json:"contentType"`
	Text        string     `json:"text"`
	FormItems   []FormItem `json:"formItems"`
	SizeBytes   int64      `json:"sizeBytes"`
	Truncated   bool       `json:"truncated"`
}

type PreparedRequest struct {
	Method         string       `json:"method"`
	URL            string       `json:"url"`
	Headers        []KeyValue   `json:"headers"`
	Cookies        []Cookie     `json:"cookies"`
	Body           PreparedBody `json:"body"`
	Proxy          ProxyConfig  `json:"proxy"`
	ProxyApplied   bool         `json:"proxyApplied"`
	ProxyExcluded  bool         `json:"proxyExcluded"`
	ProxySource    string       `json:"proxySource"`
	VariableErrors []string     `json:"variableErrors"`
	Error          string       `json:"error"`
}

type Response struct {
	StatusCode   int             `json:"statusCode"`
	Status       string          `json:"status"`
	DurationMs   int64           `json:"durationMs"`
	SizeBytes    int64           `json:"sizeBytes"`
	Headers      []KeyValue      `json:"headers"`
	Cookies      []Cookie        `json:"cookies"`
	Body         string          `json:"body"`
	ContentType  string          `json:"contentType"`
	TestResults  []TestResult    `json:"testResults"`
	Error        string          `json:"error"`
	RequestedURL string          `json:"requestedUrl"`
	Request      PreparedRequest `json:"request"`
}

type TestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type HistoryItem struct {
	ID         string    `json:"id"`
	RequestID  string    `json:"requestId"`
	Name       string    `json:"name"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	StatusCode int       `json:"statusCode"`
	DurationMs int64     `json:"durationMs"`
	CreatedAt  time.Time `json:"createdAt" ts_type:"string"`
	Request    Request   `json:"request"`
	Response   Response  `json:"response"`
}

type RunnerResult struct {
	ID            string                `json:"id"`
	CollectionID  string                `json:"collectionId"`
	EnvironmentID string                `json:"environmentId"`
	Name          string                `json:"name"`
	Iterations    int                   `json:"iterations"`
	Passed        int                   `json:"passed"`
	Failed        int                   `json:"failed"`
	DurationMs    int64                 `json:"durationMs"`
	Items         []TestResult          `json:"items"`
	Details       []RunnerRequestResult `json:"details"`
	CreatedAt     time.Time             `json:"createdAt" ts_type:"string"`
}

type RunnerRequestResult struct {
	ID          string          `json:"id"`
	RequestID   string          `json:"requestId"`
	Iteration   int             `json:"iteration"`
	Name        string          `json:"name"`
	Method      string          `json:"method"`
	URL         string          `json:"url"`
	Status      string          `json:"status"`
	StatusCode  int             `json:"statusCode"`
	DurationMs  int64           `json:"durationMs"`
	Message     string          `json:"message"`
	Request     PreparedRequest `json:"request"`
	Response    Response        `json:"response"`
	TestResults []TestResult    `json:"testResults"`
	StartedAt   time.Time       `json:"startedAt" ts_type:"string"`
	FinishedAt  time.Time       `json:"finishedAt" ts_type:"string"`
}

type VariableDebugItem struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Type     string `json:"type"`
	Raw      string `json:"raw"`
	Value    string `json:"value"`
	Resolved bool   `json:"resolved"`
	Error    string `json:"error"`
}

type VariableDebugReport struct {
	Variables []VariableDebugItem `json:"variables"`
	Errors    []string            `json:"errors"`
}

type OpenAPIImportOptions struct {
	ServerURL string `json:"serverUrl"`
}

type OpenAPIInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Servers     []string `json:"servers"`
}

type WorkspaceState struct {
	Collections         []Collection   `json:"collections"`
	Environments        []Environment  `json:"environments"`
	History             []HistoryItem  `json:"history"`
	RunnerHistory       []RunnerResult `json:"runnerHistory"`
	Globals             []KeyValue     `json:"globals"`
	Cookies             []Cookie       `json:"cookies"`
	ActiveEnvironmentID string         `json:"activeEnvironmentId"`
	Settings            Settings       `json:"settings"`
}

type Settings struct {
	Language     string      `json:"language"`
	Theme        string      `json:"theme"`
	DefaultProxy ProxyConfig `json:"defaultProxy"`
}
