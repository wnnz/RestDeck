package request

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"restdeck/internal/domain"
)

var openAPIHTTPMethods = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true, "delete": true, "head": true, "options": true,
}

type openAPIDoc struct {
	OpenAPI  string                    `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Swagger  string                    `json:"swagger,omitempty" yaml:"swagger,omitempty"`
	Host     string                    `json:"host,omitempty" yaml:"host,omitempty"`
	BasePath string                    `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	Schemes  []string                  `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Info     openAPIInfo               `json:"info" yaml:"info"`
	Servers  []openAPIServer           `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths    map[string]map[string]any `json:"paths" yaml:"paths"`
}

type openAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type openAPIServer struct {
	URL string `json:"url" yaml:"url"`
}

func ImportOpenAPI(raw string) (domain.Collection, error) {
	return ImportOpenAPIWithOptions(raw, domain.OpenAPIImportOptions{})
}

func ImportOpenAPIWithOptions(raw string, options domain.OpenAPIImportOptions) (domain.Collection, error) {
	var doc openAPIDoc
	if err := decodeOpenAPI(raw, &doc); err != nil {
		return domain.Collection{}, err
	}
	if len(doc.Paths) == 0 {
		return domain.Collection{}, fmt.Errorf("OpenAPI paths 不能为空")
	}
	now := time.Now()
	collection := domain.Collection{
		ID:          uuid.NewString(),
		Name:        fallback(doc.Info.Title, "OpenAPI Collection"),
		Description: doc.Info.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	baseURL := ""
	if strings.TrimSpace(options.ServerURL) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(options.ServerURL), "/")
	} else if len(doc.Servers) > 0 {
		baseURL = strings.TrimRight(doc.Servers[0].URL, "/")
	} else if doc.Swagger != "" {
		baseURL = strings.TrimRight(swaggerBaseURL(doc), "/")
	}

	paths := make([]string, 0, len(doc.Paths))
	for path := range doc.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	sortOrder := 0
	for _, path := range paths {
		pathItem := doc.Paths[path]
		pathParameters := arrayOfMaps(pathItem["parameters"])
		for _, method := range sortedOpenAPIMethods(pathItem) {
			operation, _ := pathItem[method].(map[string]any)
			request := importOpenAPIOperation(collection.ID, baseURL, path, method, operation, pathParameters, sortOrder)
			collection.Requests = append(collection.Requests, request)
			sortOrder++
		}
	}
	return collection, nil
}

func InspectOpenAPI(raw string) (domain.OpenAPIInfo, error) {
	var doc openAPIDoc
	if err := decodeOpenAPI(raw, &doc); err != nil {
		return domain.OpenAPIInfo{}, err
	}
	info := domain.OpenAPIInfo{
		Title:       fallback(doc.Info.Title, "OpenAPI Collection"),
		Description: doc.Info.Description,
		Servers:     []string{},
	}
	for _, server := range doc.Servers {
		if strings.TrimSpace(server.URL) != "" {
			info.Servers = append(info.Servers, server.URL)
		}
	}
	if len(info.Servers) == 0 && doc.Swagger != "" {
		info.Servers = swaggerServers(doc)
	}
	return info, nil
}

func decodeOpenAPI(raw string, doc *openAPIDoc) error {
	if err := json.Unmarshal([]byte(raw), doc); err == nil && (doc.OpenAPI != "" || doc.Swagger != "" || len(doc.Paths) > 0) {
		return nil
	}
	if err := yaml.Unmarshal([]byte(raw), doc); err != nil {
		return fmt.Errorf("OpenAPI 解析失败: %w", err)
	}
	return nil
}

func ExportOpenAPI(collection domain.Collection) (string, error) {
	doc := openAPIDoc{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       fallback(collection.Name, "RestDeck API"),
			Version:     "1.0.0",
			Description: collection.Description,
		},
		Servers: []openAPIServer{{URL: "{{baseUrl}}"}},
		Paths:   map[string]map[string]any{},
	}
	for _, req := range collection.Requests {
		path := exportOpenAPIPath(req.URL)
		method := strings.ToLower(fallback(req.Method, "GET"))
		if !openAPIHTTPMethods[method] {
			method = "get"
		}
		if doc.Paths[path] == nil {
			doc.Paths[path] = map[string]any{}
		}
		doc.Paths[path][method] = exportOpenAPIOperation(req)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func sortedOpenAPIMethods(pathItem map[string]any) []string {
	methods := []string{}
	for key := range pathItem {
		key = strings.ToLower(key)
		if openAPIHTTPMethods[key] {
			methods = append(methods, key)
		}
	}
	order := map[string]int{"get": 0, "post": 1, "put": 2, "patch": 3, "delete": 4, "head": 5, "options": 6}
	sort.Slice(methods, func(i, j int) bool { return order[methods[i]] < order[methods[j]] })
	return methods
}

func importOpenAPIOperation(collectionID, baseURL, path, method string, operation map[string]any, pathParameters []map[string]any, sortOrder int) domain.Request {
	name := stringValue(operation, "summary")
	if name == "" {
		name = stringValue(operation, "operationId")
	}
	if name == "" {
		name = strings.ToUpper(method) + " " + path
	}
	request := domain.Request{
		ID:           uuid.NewString(),
		CollectionID: collectionID,
		Name:         name,
		Method:       strings.ToUpper(method),
		URL:          baseURL + openAPIPathToTemplate(path),
		Params:       []domain.KeyValue{},
		Headers:      []domain.KeyValue{},
		BodyMode:     domain.BodyModeNone,
		Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
		Proxy:        domain.ProxyConfig{Mode: "inherit"},
		TimeoutMs:    30000,
		SortOrder:    sortOrder,
		UpdatedAt:    time.Now(),
	}
	for _, parameter := range append(pathParameters, arrayOfMaps(operation["parameters"])...) {
		target := stringValue(parameter, "in")
		if target == "body" {
			importSwaggerBodyParameter(&request, parameter)
			continue
		}
		if target == "formData" {
			importSwaggerFormParameter(&request, parameter)
			continue
		}
		schema := mapValue(parameter, "schema")
		if len(schema) == 0 {
			schema = parameter
		}
		row := domain.KeyValue{
			ID:          uuid.NewString(),
			Enabled:     true,
			Key:         stringValue(parameter, "name"),
			Value:       sampleFromSchema(schema),
			Description: stringValue(parameter, "description"),
		}
		switch target {
		case "query":
			request.Params = append(request.Params, row)
		case "header":
			request.Headers = append(request.Headers, row)
		case "path":
			if row.Value == "" {
				row.Value = "{{" + row.Key + "}}"
			}
			request.Params = append(request.Params, row)
		}
	}
	importOpenAPIRequestBody(&request, mapValue(operation, "requestBody"))
	return request
}

func importSwaggerBodyParameter(request *domain.Request, parameter map[string]any) {
	schema := mapValue(parameter, "schema")
	if len(schema) == 0 {
		return
	}
	request.BodyMode = domain.BodyModeJSON
	request.Body = jsonSampleString(sampleFromSchemaValue(schema))
}

func importSwaggerFormParameter(request *domain.Request, parameter map[string]any) {
	itemType := "text"
	if stringValue(parameter, "type") == "file" || stringValue(parameter, "format") == "binary" {
		itemType = "file"
		request.BodyMode = domain.BodyModeForm
	} else if request.BodyMode != domain.BodyModeForm {
		request.BodyMode = domain.BodyModeURLEncoded
	}
	request.FormItems = append(request.FormItems, domain.FormItem{
		ID:          uuid.NewString(),
		Enabled:     true,
		Key:         stringValue(parameter, "name"),
		Type:        itemType,
		Value:       sampleFromSchema(parameter),
		Description: stringValue(parameter, "description"),
	})
	if request.BodyMode == domain.BodyModeForm {
		request.Body = formItemsToBody(request.FormItems)
		return
	}
	request.Body = urlEncodedBodyFromFormItems(request.FormItems)
}

func importOpenAPIRequestBody(request *domain.Request, requestBody map[string]any) {
	content := mapValue(requestBody, "content")
	if len(content) == 0 {
		return
	}
	if media, ok := content["application/json"]; ok {
		request.BodyMode = domain.BodyModeJSON
		request.Body = sampleFromMedia(media)
		return
	}
	if media, ok := content["multipart/form-data"]; ok {
		request.BodyMode = domain.BodyModeForm
		request.FormItems = formItemsFromOpenAPISchema(media)
		request.Body = formItemsToBody(request.FormItems)
		return
	}
	if media, ok := content["application/x-www-form-urlencoded"]; ok {
		request.BodyMode = domain.BodyModeURLEncoded
		request.Body = kvBodyFromSchema(media)
	}
}

func sampleFromMedia(media any) string {
	mediaMap, _ := media.(map[string]any)
	if example, ok := mediaMap["example"]; ok {
		return jsonSampleString(example)
	}
	examples := mapValue(mediaMap, "examples")
	for _, value := range examples {
		if exampleValue := mapValue(value, "value"); len(exampleValue) > 0 {
			return jsonSampleString(exampleValue)
		}
	}
	if sample := sampleFromSchemaValue(mediaMap["schema"]); sample != nil {
		return jsonSampleString(sample)
	}
	return "{\n  \n}"
}

func formItemsFromOpenAPISchema(media any) []domain.FormItem {
	mediaMap, _ := media.(map[string]any)
	schema := mapValue(mediaMap, "schema")
	properties := mapValue(schema, "properties")
	items := []domain.FormItem{}
	keys := sortedKeys(properties)
	for _, key := range keys {
		prop := mapValue(properties, key)
		itemType := "text"
		if stringValue(prop, "format") == "binary" || stringValue(prop, "type") == "file" {
			itemType = "file"
		}
		items = append(items, domain.FormItem{
			ID:          uuid.NewString(),
			Enabled:     true,
			Key:         key,
			Type:        itemType,
			Value:       sampleFromSchema(prop),
			Description: stringValue(prop, "description"),
		})
	}
	return items
}

func kvBodyFromSchema(media any) string {
	lines := []string{}
	for _, item := range formItemsFromOpenAPISchema(media) {
		if item.Type == "file" {
			continue
		}
		lines = append(lines, item.Key+"="+item.Value)
	}
	return strings.Join(lines, "\n")
}

func swaggerServers(doc openAPIDoc) []string {
	basePath := strings.TrimSpace(doc.BasePath)
	if basePath == "" {
		basePath = "/"
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	host := strings.TrimSpace(doc.Host)
	if host == "" {
		return []string{strings.TrimRight(basePath, "/")}
	}
	schemes := doc.Schemes
	if len(schemes) == 0 {
		schemes = []string{"https"}
	}
	servers := []string{}
	for _, scheme := range schemes {
		scheme = strings.TrimSpace(scheme)
		if scheme == "" {
			continue
		}
		servers = append(servers, strings.TrimRight(scheme+"://"+host+basePath, "/"))
	}
	return servers
}

func swaggerBaseURL(doc openAPIDoc) string {
	servers := swaggerServers(doc)
	if len(servers) == 0 {
		return ""
	}
	return servers[0]
}

func urlEncodedBodyFromFormItems(items []domain.FormItem) string {
	lines := []string{}
	for _, item := range items {
		if !item.Enabled || item.Key == "" || item.Type == "file" {
			continue
		}
		lines = append(lines, item.Key+"="+item.Value)
	}
	return strings.Join(lines, "\n")
}

func exportOpenAPIOperation(req domain.Request) map[string]any {
	operation := map[string]any{
		"summary":     fallback(req.Name, req.Method+" "+req.URL),
		"description": "",
		"responses": map[string]any{
			"200": map[string]any{"description": "OK"},
		},
	}
	parameters := []map[string]any{}
	for _, param := range req.Params {
		if param.Enabled && param.Key != "" {
			parameters = append(parameters, exportOpenAPIParameter(param, "query"))
		}
	}
	for _, header := range req.Headers {
		if header.Enabled && header.Key != "" && !strings.EqualFold(header.Key, "Content-Type") {
			parameters = append(parameters, exportOpenAPIParameter(header, "header"))
		}
	}
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}
	if body := exportOpenAPIRequestBody(req); len(body) > 0 {
		operation["requestBody"] = body
	}
	return operation
}

func exportOpenAPIParameter(kv domain.KeyValue, target string) map[string]any {
	return map[string]any{
		"name":        kv.Key,
		"in":          target,
		"description": kv.Description,
		"required":    false,
		"schema":      map[string]any{"type": "string", "example": kv.Value},
	}
}

func exportOpenAPIRequestBody(req domain.Request) map[string]any {
	switch req.BodyMode {
	case domain.BodyModeJSON:
		var body any
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			body = req.Body
		}
		return map[string]any{
			"content": map[string]any{
				"application/json": map[string]any{"example": body},
			},
		}
	case domain.BodyModeRaw:
		return map[string]any{
			"content": map[string]any{
				"text/plain": map[string]any{"example": req.Body},
			},
		}
	case domain.BodyModeForm:
		properties := map[string]any{}
		for _, item := range req.FormItems {
			if !item.Enabled || item.Key == "" {
				continue
			}
			prop := map[string]any{"type": "string", "description": item.Description}
			if item.Type == "file" {
				prop["format"] = "binary"
			} else {
				prop["example"] = item.Value
			}
			properties[item.Key] = prop
		}
		return map[string]any{
			"content": map[string]any{
				"multipart/form-data": map[string]any{
					"schema": map[string]any{"type": "object", "properties": properties},
				},
			},
		}
	case domain.BodyModeURLEncoded:
		properties := map[string]any{}
		for _, kv := range linesToKV(req.Body) {
			if kv.Key != "" {
				properties[kv.Key] = map[string]any{"type": "string", "example": kv.Value}
			}
		}
		return map[string]any{
			"content": map[string]any{
				"application/x-www-form-urlencoded": map[string]any{
					"schema": map[string]any{"type": "object", "properties": properties},
				},
			},
		}
	default:
		return nil
	}
}

func exportOpenAPIPath(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return "/"
	}
	value = strings.ReplaceAll(value, "{{baseUrl}}", "")
	parsed, err := url.Parse(value)
	if err == nil && parsed.Path != "" {
		return parsed.Path
	}
	if strings.HasPrefix(value, "/") {
		if index := strings.Index(value, "?"); index >= 0 {
			return value[:index]
		}
		return value
	}
	return "/" + strings.TrimLeft(value, "/")
}

func openAPIPathToTemplate(path string) string {
	var out strings.Builder
	for index := 0; index < len(path); {
		if path[index] == '{' {
			end := strings.IndexByte(path[index+1:], '}')
			if end >= 0 {
				end += index + 1
				name := strings.TrimSpace(path[index+1 : end])
				if name != "" {
					out.WriteString("{{")
					out.WriteString(name)
					out.WriteString("}}")
					index = end + 1
					continue
				}
			}
		}
		out.WriteByte(path[index])
		index++
	}
	return out.String()
}

func arrayOfMaps(raw any) []map[string]any {
	values, _ := raw.([]any)
	out := []map[string]any{}
	for _, value := range values {
		if item, ok := value.(map[string]any); ok {
			out = append(out, item)
		}
	}
	return out
}

func mapValue(source any, key string) map[string]any {
	sourceMap, _ := source.(map[string]any)
	value, _ := sourceMap[key].(map[string]any)
	return value
}

func stringValue(source map[string]any, key string) string {
	value, _ := source[key].(string)
	return value
}

func sampleFromSchema(schema map[string]any) string {
	if example, ok := schema["example"]; ok {
		return fmt.Sprint(example)
	}
	if def, ok := schema["default"]; ok {
		return fmt.Sprint(def)
	}
	if enumValues, ok := schema["enum"].([]any); ok && len(enumValues) > 0 {
		return fmt.Sprint(enumValues[0])
	}
	switch stringValue(schema, "type") {
	case "integer", "number":
		return "0"
	case "boolean":
		return "true"
	case "array", "object":
		return jsonSampleString(sampleFromSchemaValue(schema))
	default:
		return ""
	}
}

func sampleFromSchemaValue(schema any) any {
	schemaMap, _ := schema.(map[string]any)
	if example, ok := schemaMap["example"]; ok {
		return example
	}
	if def, ok := schemaMap["default"]; ok {
		return def
	}
	if enumValues, ok := schemaMap["enum"].([]any); ok && len(enumValues) > 0 {
		return enumValues[0]
	}
	switch stringValue(schemaMap, "type") {
	case "object":
		out := map[string]any{}
		properties := mapValue(schemaMap, "properties")
		for _, key := range sortedKeys(properties) {
			out[key] = sampleFromSchemaValue(properties[key])
		}
		return out
	case "array":
		return []any{sampleFromSchemaValue(schemaMap["items"])}
	case "integer", "number":
		return 0
	case "boolean":
		return true
	case "string":
		if example, ok := schemaMap["example"]; ok {
			return example
		}
		return ""
	default:
		return map[string]any{}
	}
}

func jsonSampleString(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
