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
	Mode string `json:"mode"`
	URL  string `json:"url"`
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
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type Folder struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collectionId"`
	ParentID     string    `json:"parentId"`
	Name         string    `json:"name"`
	SortOrder    int       `json:"sortOrder"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Folders     []Folder  `json:"folders"`
	Requests    []Request `json:"requests"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Environment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Variables []KeyValue `json:"variables"`
	IsActive  bool       `json:"isActive"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	HTTPOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}

type Response struct {
	StatusCode   int          `json:"statusCode"`
	Status       string       `json:"status"`
	DurationMs   int64        `json:"durationMs"`
	SizeBytes    int64        `json:"sizeBytes"`
	Headers      []KeyValue   `json:"headers"`
	Cookies      []Cookie     `json:"cookies"`
	Body         string       `json:"body"`
	ContentType  string       `json:"contentType"`
	TestResults  []TestResult `json:"testResults"`
	Error        string       `json:"error"`
	RequestedURL string       `json:"requestedUrl"`
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
	CreatedAt  time.Time `json:"createdAt"`
	Request    Request   `json:"request"`
	Response   Response  `json:"response"`
}

type RunnerResult struct {
	ID            string       `json:"id"`
	CollectionID  string       `json:"collectionId"`
	EnvironmentID string       `json:"environmentId"`
	Name          string       `json:"name"`
	Iterations    int          `json:"iterations"`
	Passed        int          `json:"passed"`
	Failed        int          `json:"failed"`
	DurationMs    int64        `json:"durationMs"`
	Items         []TestResult `json:"items"`
	CreatedAt     time.Time    `json:"createdAt"`
}

type WorkspaceState struct {
	Collections         []Collection  `json:"collections"`
	Environments        []Environment `json:"environments"`
	History             []HistoryItem `json:"history"`
	Globals             []KeyValue    `json:"globals"`
	ActiveEnvironmentID string        `json:"activeEnvironmentId"`
	Settings            Settings      `json:"settings"`
}

type Settings struct {
	Language     string      `json:"language"`
	Theme        string      `json:"theme"`
	DefaultProxy ProxyConfig `json:"defaultProxy"`
}
