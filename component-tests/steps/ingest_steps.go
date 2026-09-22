// ingest_steps — срезовые godog-шаги slice-01-ingest-report: тело из файла-фикстуры,
// проверка принятых рёбер и графа. Механический клей (HTTP + JSON-ассерты);
// чёрный ящик — запущенный сервис, не внутренности.
package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cucumber/godog"
)

func (w *World) registerIngestSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^запущенный сервис netlist$`, w.serviceUp)
	ctx.Step(`^клиент отправляет POST /reports с телом из файла (\S+)$`, w.postReportFromFile)
	ctx.Step(`^граф содержит ребро (\S+) → (\S+) с субъектом "([^"]*)"$`, w.graphHasEdge)
}

// serviceUp — тривиальный Given: compose гарантирует healthcheck до старта раннера.
func (w *World) serviceUp() error {
	return w.doRequest(http.MethodGet, "/health", nil)
}

// postReportFromFile — POST /reports с телом из bind-mounted файла фикстуры.
func (w *World) postReportFromFile(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return w.doRequest(http.MethodPost, "/reports", body)
}

// graphHasEdge — GET /graph содержит ребро consumer→provider с данным субъектом.
func (w *World) graphHasEdge(consumer, provider, subject string) error {
	if err := w.doRequest(http.MethodGet, "/graph", nil); err != nil {
		return err
	}
	var graph struct {
		Edges []struct {
			Consumer   string `json:"consumer"`
			Provider   string `json:"provider"`
			Subject    string `json:"subject"`
			Compatible bool   `json:"compatible"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(w.lastBody, &graph); err != nil {
		return fmt.Errorf("graph is not JSON: %v; body=%s", err, w.lastBody)
	}
	for _, e := range graph.Edges {
		if e.Consumer == consumer && e.Provider == provider && e.Subject == subject {
			return nil
		}
	}
	return fmt.Errorf("edge %s → %s %q not found in graph: %s", consumer, provider, subject, w.lastBody)
}
