package request

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

type harArchive struct {
	Log harLog `json:"log"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Entries []harEntry `json:"entries"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harEntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            int64       `json:"time"`
	Request         harRequest  `json:"request"`
	Response        harResponse `json:"response"`
}

type harRequest struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []harNameVal `json:"headers"`
	QueryString []harNameVal `json:"queryString"`
	Cookies     []harNameVal `json:"cookies"`
	PostData    *harPostData `json:"postData,omitempty"`
	HeadersSize int          `json:"headersSize"`
	BodySize    int          `json:"bodySize"`
}

type harResponse struct {
	Status      int          `json:"status"`
	StatusText  string       `json:"statusText"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []harNameVal `json:"headers"`
	Cookies     []harNameVal `json:"cookies"`
	Content     harContent   `json:"content"`
	HeadersSize int          `json:"headersSize"`
	BodySize    int          `json:"bodySize"`
}

type harNameVal struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPostData struct {
	MimeType string       `json:"mimeType"`
	Text     string       `json:"text,omitempty"`
	Params   []harNameVal `json:"params,omitempty"`
}

type harContent struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

func ImportHAR(raw string) (domain.Collection, error) {
	var archive harArchive
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		return domain.Collection{}, fmt.Errorf("HAR JSON 解析失败: %w", err)
	}
	if len(archive.Log.Entries) == 0 {
		return domain.Collection{}, fmt.Errorf("HAR entries 不能为空")
	}
	now := time.Now()
	collection := domain.Collection{
		ID:        uuid.NewString(),
		Name:      "HAR Import",
		CreatedAt: now,
		UpdatedAt: now,
	}
	for index, entry := range archive.Log.Entries {
		req := domain.Request{
			ID:           uuid.NewString(),
			CollectionID: collection.ID,
			Name:         fallback(entry.Request.Method+" "+entry.Request.URL, "HAR Request"),
			Method:       fallback(entry.Request.Method, "GET"),
			URL:          entry.Request.URL,
			Params:       harPairsToKV(entry.Request.QueryString),
			Headers:      harPairsToKV(entry.Request.Headers),
			BodyMode:     domain.BodyModeNone,
			Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
			Proxy:        domain.ProxyConfig{Mode: "inherit"},
			TimeoutMs:    30000,
			SortOrder:    index,
			UpdatedAt:    now,
		}
		if entry.Request.PostData != nil {
			applyHARPostData(&req, *entry.Request.PostData)
		}
		collection.Requests = append(collection.Requests, req)
	}
	return collection, nil
}

func ExportHAR(collection domain.Collection) (string, error) {
	archive := harArchive{Log: harLog{
		Version: "1.2",
		Creator: harCreator{Name: "RestDeck", Version: "1.0"},
		Entries: []harEntry{},
	}}
	for _, req := range collection.Requests {
		entry := harEntry{
			StartedDateTime: time.Now().UTC().Format(time.RFC3339),
			Time:            0,
			Request: harRequest{
				Method:      fallback(req.Method, "GET"),
				URL:         req.URL,
				HTTPVersion: "HTTP/1.1",
				Headers:     kvToHARPairs(req.Headers),
				QueryString: kvToHARPairs(req.Params),
				Cookies:     []harNameVal{},
				HeadersSize: -1,
				BodySize:    -1,
			},
			Response: harResponse{
				Status:      0,
				StatusText:  "",
				HTTPVersion: "HTTP/1.1",
				Headers:     []harNameVal{},
				Cookies:     []harNameVal{},
				Content:     harContent{Size: 0, MimeType: ""},
				HeadersSize: -1,
				BodySize:    -1,
			},
		}
		if postData := requestToHARPostData(req); postData != nil {
			entry.Request.PostData = postData
			entry.Request.BodySize = len(postData.Text)
		}
		archive.Log.Entries = append(archive.Log.Entries, entry)
	}
	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func harPairsToKV(items []harNameVal) []domain.KeyValue {
	out := []domain.KeyValue{}
	for _, item := range items {
		out = append(out, domain.KeyValue{
			ID:      uuid.NewString(),
			Enabled: true,
			Key:     item.Name,
			Value:   item.Value,
		})
	}
	return out
}

func kvToHARPairs(items []domain.KeyValue) []harNameVal {
	out := []harNameVal{}
	for _, item := range items {
		if item.Enabled && item.Key != "" {
			out = append(out, harNameVal{Name: item.Key, Value: item.Value})
		}
	}
	return out
}

func applyHARPostData(req *domain.Request, postData harPostData) {
	mimeType := strings.ToLower(postData.MimeType)
	switch {
	case strings.Contains(mimeType, "application/json"):
		req.BodyMode = domain.BodyModeJSON
		req.Body = decodeHARText(postData.Text)
	case strings.Contains(mimeType, "application/x-www-form-urlencoded"):
		req.BodyMode = domain.BodyModeURLEncoded
		if len(postData.Params) > 0 {
			req.Body = harParamsToBody(postData.Params)
		} else {
			req.Body = decodeHARText(postData.Text)
		}
	case strings.Contains(mimeType, "multipart/form-data"):
		req.BodyMode = domain.BodyModeForm
		for _, param := range postData.Params {
			req.FormItems = append(req.FormItems, domain.FormItem{
				ID:      uuid.NewString(),
				Enabled: true,
				Key:     param.Name,
				Type:    "text",
				Value:   param.Value,
			})
		}
		req.Body = formItemsToBody(req.FormItems)
	default:
		req.BodyMode = domain.BodyModeRaw
		req.Body = decodeHARText(postData.Text)
	}
}

func requestToHARPostData(req domain.Request) *harPostData {
	switch req.BodyMode {
	case domain.BodyModeJSON:
		return &harPostData{MimeType: "application/json", Text: req.Body}
	case domain.BodyModeRaw:
		return &harPostData{MimeType: "text/plain", Text: req.Body}
	case domain.BodyModeURLEncoded:
		return &harPostData{MimeType: "application/x-www-form-urlencoded", Text: req.Body, Params: bodyLinesToHARParams(req.Body)}
	case domain.BodyModeForm:
		params := []harNameVal{}
		for _, item := range req.FormItems {
			if item.Enabled && item.Key != "" {
				value := item.Value
				if item.Type == "file" {
					value = "@" + item.FilePath
				}
				params = append(params, harNameVal{Name: item.Key, Value: value})
			}
		}
		return &harPostData{MimeType: "multipart/form-data", Text: formItemsToBody(req.FormItems), Params: params}
	default:
		return nil
	}
}

func harParamsToBody(items []harNameVal) string {
	lines := []string{}
	for _, item := range items {
		lines = append(lines, item.Name+"="+item.Value)
	}
	return strings.Join(lines, "\n")
}

func bodyLinesToHARParams(raw string) []harNameVal {
	out := []harNameVal{}
	for _, kv := range linesToKV(raw) {
		out = append(out, harNameVal{Name: kv.Key, Value: kv.Value})
	}
	return out
}

func decodeHARText(text string) string {
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err == nil && len(decoded) > 0 && !strings.ContainsRune(string(decoded), '\x00') {
		return string(decoded)
	}
	return text
}
