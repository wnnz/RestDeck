package request

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

type postmanCollection struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      string `json:"schema"`
	} `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Item    []postmanItem   `json:"item,omitempty"`
	Request *postmanRequest `json:"request,omitempty"`
}

type postmanRequest struct {
	Method string          `json:"method"`
	Header []postmanKV     `json:"header"`
	URL    json.RawMessage `json:"url"`
	Body   *postmanBody    `json:"body,omitempty"`
	Auth   *postmanAuth    `json:"auth,omitempty"`
	Event  []postmanEvent  `json:"event,omitempty"`
}

type postmanKV struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
	Type        string `json:"type"`
}

type postmanBody struct {
	Mode       string      `json:"mode"`
	Raw        string      `json:"raw"`
	URLEncoded []postmanKV `json:"urlencoded"`
	FormData   []postmanKV `json:"formdata"`
}

type postmanAuth struct {
	Type   string      `json:"type"`
	APIKey []postmanKV `json:"apikey"`
	Bearer []postmanKV `json:"bearer"`
	Basic  []postmanKV `json:"basic"`
	Digest []postmanKV `json:"digest"`
	OAuth1 []postmanKV `json:"oauth1"`
	OAuth2 []postmanKV `json:"oauth2"`
}

type postmanEvent struct {
	Listen string `json:"listen"`
	Script struct {
		Exec []string `json:"exec"`
		Type string   `json:"type"`
	} `json:"script"`
}

func ImportPostman(raw string) (domain.Collection, error) {
	var pc postmanCollection
	if err := json.Unmarshal([]byte(raw), &pc); err != nil {
		return domain.Collection{}, err
	}
	now := time.Now()
	c := domain.Collection{
		ID:          uuid.NewString(),
		Name:        fallback(pc.Info.Name, "Imported Collection"),
		Description: pc.Info.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	walkPostmanItems(&c, "", pc.Item)
	return c, nil
}

func ExportPostman(collection domain.Collection) (string, error) {
	pc := postmanCollection{}
	pc.Info.Name = collection.Name
	pc.Info.Description = collection.Description
	pc.Info.Schema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	pc.Item = buildPostmanTree(collection, "")
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func walkPostmanItems(c *domain.Collection, parentID string, items []postmanItem) {
	for index, item := range items {
		if item.Request == nil {
			folderID := uuid.NewString()
			c.Folders = append(c.Folders, domain.Folder{
				ID:           folderID,
				CollectionID: c.ID,
				ParentID:     parentID,
				Name:         fallback(item.Name, "Folder"),
				SortOrder:    index,
				UpdatedAt:    time.Now(),
			})
			walkPostmanItems(c, folderID, item.Item)
			continue
		}
		req := domain.Request{
			ID:           uuid.NewString(),
			CollectionID: c.ID,
			ParentID:     parentID,
			Name:         fallback(item.Name, "Request"),
			Method:       fallback(item.Request.Method, "GET"),
			URL:          parsePostmanURL(item.Request.URL),
			Headers:      importKV(item.Request.Header),
			BodyMode:     domain.BodyModeNone,
			Auth:         importAuth(item.Request.Auth),
			TimeoutMs:    30000,
			SortOrder:    index,
			UpdatedAt:    time.Now(),
		}
		if item.Request.Body != nil {
			req.BodyMode, req.Body = importBody(*item.Request.Body)
		}
		for _, event := range item.Request.Event {
			script := strings.Join(event.Script.Exec, "\n")
			switch event.Listen {
			case "prerequest":
				req.PreScript = script
			case "test":
				req.TestScript = script
			}
		}
		c.Requests = append(c.Requests, req)
	}
}

func parsePostmanURL(raw json.RawMessage) string {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}
	var obj struct {
		Raw      string      `json:"raw"`
		Protocol string      `json:"protocol"`
		Host     []string    `json:"host"`
		Path     []string    `json:"path"`
		Query    []postmanKV `json:"query"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	if obj.Raw != "" {
		return obj.Raw
	}
	url := ""
	if obj.Protocol != "" {
		url = obj.Protocol + "://"
	}
	url += strings.Join(obj.Host, ".")
	if len(obj.Path) > 0 {
		url += "/" + strings.Join(obj.Path, "/")
	}
	return url
}

func importKV(items []postmanKV) []domain.KeyValue {
	out := []domain.KeyValue{}
	for _, item := range items {
		out = append(out, domain.KeyValue{
			ID:          uuid.NewString(),
			Enabled:     !item.Disabled,
			Key:         item.Key,
			Value:       item.Value,
			Description: item.Description,
		})
	}
	return out
}

func importBody(body postmanBody) (domain.BodyMode, string) {
	switch body.Mode {
	case "raw":
		if strings.TrimSpace(body.Raw) == "" {
			return domain.BodyModeRaw, ""
		}
		if strings.HasPrefix(strings.TrimSpace(body.Raw), "{") || strings.HasPrefix(strings.TrimSpace(body.Raw), "[") {
			return domain.BodyModeJSON, body.Raw
		}
		return domain.BodyModeRaw, body.Raw
	case "urlencoded":
		return domain.BodyModeURLEncoded, kvLines(body.URLEncoded)
	case "formdata":
		return domain.BodyModeForm, kvLines(body.FormData)
	default:
		return domain.BodyModeNone, ""
	}
}

func kvLines(items []postmanKV) string {
	lines := []string{}
	for _, item := range items {
		if !item.Disabled {
			lines = append(lines, item.Key+"="+item.Value)
		}
	}
	return strings.Join(lines, "\n")
}

func importAuth(auth *postmanAuth) domain.AuthConfig {
	if auth == nil || auth.Type == "" {
		return domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}}
	}
	values := map[string]string{}
	for _, kv := range authValues(auth) {
		values[kv.Key] = kv.Value
	}
	switch auth.Type {
	case "apikey":
		return domain.AuthConfig{Type: domain.AuthTypeAPIKey, Values: map[string]string{"key": values["key"], "value": values["value"], "in": values["in"]}}
	case "bearer":
		return domain.AuthConfig{Type: domain.AuthTypeBearer, Values: map[string]string{"token": values["token"]}}
	case "basic":
		return domain.AuthConfig{Type: domain.AuthTypeBasic, Values: map[string]string{"username": values["username"], "password": values["password"]}}
	case "digest":
		return domain.AuthConfig{Type: domain.AuthTypeDigest, Values: map[string]string{"username": values["username"], "password": values["password"]}}
	case "oauth1":
		return domain.AuthConfig{Type: domain.AuthTypeOAuth1, Values: map[string]string{
			"consumerKey": values["consumerKey"], "consumerSecret": values["consumerSecret"], "token": values["token"], "tokenSecret": values["tokenSecret"],
		}}
	case "oauth2":
		return domain.AuthConfig{Type: domain.AuthTypeOAuth2, Values: map[string]string{"accessToken": values["accessToken"]}}
	default:
		return domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}}
	}
}

func authValues(auth *postmanAuth) []postmanKV {
	switch auth.Type {
	case "apikey":
		return auth.APIKey
	case "bearer":
		return auth.Bearer
	case "basic":
		return auth.Basic
	case "digest":
		return auth.Digest
	case "oauth1":
		return auth.OAuth1
	case "oauth2":
		return auth.OAuth2
	default:
		return nil
	}
}

func buildPostmanTree(collection domain.Collection, parentID string) []postmanItem {
	items := []postmanItem{}
	for _, folder := range collection.Folders {
		if folder.ParentID == parentID {
			items = append(items, postmanItem{Name: folder.Name, Item: buildPostmanTree(collection, folder.ID)})
		}
	}
	for _, req := range collection.Requests {
		if req.ParentID == parentID {
			items = append(items, exportRequest(req))
		}
	}
	return items
}

func exportRequest(req domain.Request) postmanItem {
	body := exportBody(req)
	pmReq := &postmanRequest{
		Method: req.Method,
		Header: exportKV(req.Headers),
		URL:    json.RawMessage(fmt.Sprintf("%q", req.URL)),
		Body:   body,
		Auth:   exportAuth(req.Auth),
	}
	if req.PreScript != "" {
		pmReq.Event = append(pmReq.Event, postmanEvent{Listen: "prerequest"})
		pmReq.Event[len(pmReq.Event)-1].Script.Exec = strings.Split(req.PreScript, "\n")
		pmReq.Event[len(pmReq.Event)-1].Script.Type = "text/javascript"
	}
	if req.TestScript != "" {
		pmReq.Event = append(pmReq.Event, postmanEvent{Listen: "test"})
		pmReq.Event[len(pmReq.Event)-1].Script.Exec = strings.Split(req.TestScript, "\n")
		pmReq.Event[len(pmReq.Event)-1].Script.Type = "text/javascript"
	}
	return postmanItem{Name: req.Name, Request: pmReq}
}

func exportKV(items []domain.KeyValue) []postmanKV {
	out := []postmanKV{}
	for _, item := range items {
		out = append(out, postmanKV{Key: item.Key, Value: item.Value, Description: item.Description, Disabled: !item.Enabled})
	}
	return out
}

func exportBody(req domain.Request) *postmanBody {
	switch req.BodyMode {
	case domain.BodyModeJSON:
		return &postmanBody{Mode: "raw", Raw: req.Body}
	case domain.BodyModeRaw:
		return &postmanBody{Mode: "raw", Raw: req.Body}
	case domain.BodyModeURLEncoded:
		return &postmanBody{Mode: "urlencoded", URLEncoded: linesToKV(req.Body)}
	case domain.BodyModeForm:
		return &postmanBody{Mode: "formdata", FormData: linesToKV(req.Body)}
	default:
		return nil
	}
}

func linesToKV(raw string) []postmanKV {
	out := []postmanKV{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		out = append(out, postmanKV{Key: parts[0], Value: value})
	}
	return out
}

func exportAuth(auth domain.AuthConfig) *postmanAuth {
	if auth.Type == "" || auth.Type == domain.AuthTypeNone {
		return nil
	}
	pm := &postmanAuth{}
	switch auth.Type {
	case domain.AuthTypeAPIKey:
		pm.Type = "apikey"
		pm.APIKey = []postmanKV{{Key: "key", Value: auth.Values["key"]}, {Key: "value", Value: auth.Values["value"]}, {Key: "in", Value: auth.Values["in"]}}
	case domain.AuthTypeBearer:
		pm.Type = "bearer"
		pm.Bearer = []postmanKV{{Key: "token", Value: auth.Values["token"]}}
	case domain.AuthTypeBasic:
		pm.Type = "basic"
		pm.Basic = []postmanKV{{Key: "username", Value: auth.Values["username"]}, {Key: "password", Value: auth.Values["password"]}}
	case domain.AuthTypeDigest:
		pm.Type = "digest"
		pm.Digest = []postmanKV{{Key: "username", Value: auth.Values["username"]}, {Key: "password", Value: auth.Values["password"]}}
	case domain.AuthTypeOAuth1:
		pm.Type = "oauth1"
		pm.OAuth1 = []postmanKV{{Key: "consumerKey", Value: auth.Values["consumerKey"]}, {Key: "consumerSecret", Value: auth.Values["consumerSecret"]}, {Key: "token", Value: auth.Values["token"]}, {Key: "tokenSecret", Value: auth.Values["tokenSecret"]}}
	case domain.AuthTypeOAuth2:
		pm.Type = "oauth2"
		pm.OAuth2 = []postmanKV{{Key: "accessToken", Value: auth.Values["accessToken"]}}
	}
	return pm
}

func fallback(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}
