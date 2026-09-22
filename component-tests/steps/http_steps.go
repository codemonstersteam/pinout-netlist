package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
)

// registerHTTPSteps — универсальные HTTP-степы (чёрный ящик по HTTP):
// отправка запроса (с/без тела) и проверки ответа (статус, заголовок, JSON-поля).
// Достаточно для большинства компонентных сценариев против OpenAPI-схемы.
func (w *World) registerHTTPSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^клиент отправляет (GET|POST|PUT|DELETE) (\S+)$`, w.sendRequest)
	ctx.Step(`^клиент отправляет (GET|POST|PUT|DELETE) (\S+) с телом:$`, w.sendRequestWithBody)
	ctx.Step(`^ответ (\d+)(?:\s|$)`, w.responseStatus)
	ctx.Step(`^ответ содержит заголовок ([A-Za-z\-]+)$`, w.responseHasHeader)
	ctx.Step(`^ответ содержит JSON-поле ([\w\.\[\]]+) со значением "([^"]*)"$`, w.responseJSONField)
	ctx.Step(`^ответ содержит непустое JSON-поле ([\w\.\[\]]+)$`, w.responseJSONFieldNonEmpty)
	ctx.Step(`^ответ содержит JSON-поле ([\w\.\[\]]+)$`, w.responseJSONFieldPresent)
}

func (w *World) sendRequest(method, path string) error { return w.doRequest(method, path, nil) }

func (w *World) sendRequestWithBody(method, path string, body *godog.DocString) error {
	return w.doRequest(method, path, []byte(body.Content))
}

func (w *World) doRequest(method, path string, body []byte) error {
	url := w.serviceBaseURL + path
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if w.lastResponse != nil {
		_ = w.lastResponse.Body.Close()
	}
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send %s %s: %w", method, url, err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("read body: %w", err)
	}
	w.lastResponse = resp
	w.lastBody = respBody
	return nil
}

func (w *World) responseStatus(expected int) error {
	if w.lastResponse == nil {
		return fmt.Errorf("ответ ещё не получен")
	}
	if w.lastResponse.StatusCode != expected {
		return fmt.Errorf("ожидали статус %d, получили %d (тело: %s)",
			expected, w.lastResponse.StatusCode, string(w.lastBody))
	}
	return nil
}

func (w *World) responseHasHeader(name string) error {
	if w.lastResponse == nil {
		return fmt.Errorf("ответ ещё не получен")
	}
	if w.lastResponse.Header.Get(name) == "" {
		return fmt.Errorf("ожидали заголовок %q, не нашли", name)
	}
	return nil
}

func (w *World) responseJSONFieldPresent(field string) error {
	if _, ok := w.jsonPath(field); !ok {
		return fmt.Errorf("в ответе нет поля %q (тело: %s)", field, string(w.lastBody))
	}
	return nil
}

func (w *World) responseJSONFieldNonEmpty(field string) error {
	v, ok := w.jsonPath(field)
	if !ok {
		return fmt.Errorf("в ответе нет поля %q", field)
	}
	if fmt.Sprintf("%v", v) == "" {
		return fmt.Errorf("поле %q пустое", field)
	}
	return nil
}

func (w *World) responseJSONField(field, expected string) error {
	v, ok := w.jsonPath(field)
	if !ok {
		return fmt.Errorf("в ответе нет поля %q (тело: %s)", field, string(w.lastBody))
	}
	if got := fmt.Sprintf("%v", v); got != expected {
		return fmt.Errorf("поле %q: ожидали %q, получили %q", field, expected, got)
	}
	return nil
}

// jsonPath — чтение (возможно вложенного) поля по точечному пути: error.code.
func (w *World) jsonPath(path string) (any, bool) {
	m, err := w.parseJSON()
	if err != nil {
		return nil, false
	}
	var cur any = m
	for _, part := range strings.Split(path, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = obj[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func (w *World) parseJSON() (map[string]any, error) {
	if w.lastBody == nil {
		return nil, fmt.Errorf("ответ ещё не получен")
	}
	var parsed map[string]any
	if err := json.Unmarshal(w.lastBody, &parsed); err != nil {
		return nil, fmt.Errorf("ответ не валидный JSON: %w (тело: %s)", err, string(w.lastBody))
	}
	return parsed, nil
}
