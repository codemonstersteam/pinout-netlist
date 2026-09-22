// Package steps — godog-степы компонентных тестов сервиса.
// World — состояние сценария (HTTP-клиент + последний ответ), создаётся заново
// на каждый Scenario. Доменных зависимостей нет — добавляй свои фикстуры/степы
// в новых *_steps.go рядом.
package steps

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/cucumber/godog"
)

type World struct {
	serviceBaseURL string
	httpClient     *http.Client

	lastResponse *http.Response
	lastBody     []byte
}

func newWorld() *World {
	return &World{
		serviceBaseURL: getenv("SERVICE_BASE_URL", "http://service:8080"),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *World) resetState() {
	w.lastResponse = nil
	w.lastBody = nil
}

func (w *World) beforeScenario(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	w.resetState()
	return ctx, nil
}

func (w *World) afterScenario(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if w.lastResponse != nil {
		_ = w.lastResponse.Body.Close()
	}
	return ctx, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
