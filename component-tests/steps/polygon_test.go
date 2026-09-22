// polygon_steps + TestPolygon — E2E-полигон тройки pinout (netlist + реальные
// бинари обоих валидаторов). Отдельный godog-сьют (features/polygon.feature),
// запускается run-polygon.sh (GO_TEST_RUN=TestPolygon): обычным компонентным
// прогонам бинари валидаторов не нужны.
package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/cucumber/godog"
)

// runValidatorAndIngest — прогнать бинарь валидатора на конфиге, взять отчёт
// (stdout, JSON канона 1.1), отправить в netlist конвертом {report, subjects}
// (subjects — scope потребителя из его же конфига: operations / channels).
func (w *World) runValidatorAndIngest(tool, configPath string) error {
	bin := os.Getenv(tool)
	if bin == "" {
		return fmt.Errorf("env %s не задан (полигон запускается run-polygon.sh)", tool)
	}
	cmd := exec.Command(bin, "validate", configPath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// exit 1 = verdict «несовместимо» — это валидный результат, не сбой обвязки.
		if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() > 1 {
			return fmt.Errorf("%s validate: %v; stdout=%s", tool, err, stdout.String())
		}
	}
	report := stdout.Bytes()
	if !json.Valid(report) {
		return fmt.Errorf("%s stdout не JSON: %.200s", tool, report)
	}

	var body bytes.Buffer
	body.WriteString(`{"report":`)
	body.Write(report)
	if subjects := subjectsFromConfig(configPath); len(subjects) > 0 {
		enc := make([]string, 0, len(subjects))
		for _, s := range subjects {
			b, _ := json.Marshal(s)
			enc = append(enc, string(b))
		}
		body.WriteString(`,"subjects":[` + strings.Join(enc, ",") + `]`)
	}
	body.WriteString(`}`)

	return w.doRequest(http.MethodPost, "/reports", body.Bytes())
}

// subjectsFromConfig — перечень субъектов потребителя из его же конфига
// (consumer.operations → `METHOD /path`; consumer.channels → адреса): именно
// этот scope отсутствует в совместимом отчёте (TASK.md экосистемы, полигон).
func subjectsFromConfig(configPath string) []string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}
	var cfg struct {
		Consumer struct {
			Operations []struct {
				Path   string `yaml:"path"`
				Method string `yaml:"method"`
			} `yaml:"operations"`
			Channels []string `yaml:"channels"`
		} `yaml:"consumer"`
	}
	if yaml.Unmarshal(data, &cfg) != nil {
		return nil
	}
	var out []string
	for _, op := range cfg.Consumer.Operations {
		out = append(out, strings.ToUpper(op.Method)+" "+op.Path)
	}
	return append(out, cfg.Consumer.Channels...)
}

// graphEdge — найти рёбра графа по consumer/provider/subject.
func (w *World) graphEdges() ([]map[string]any, error) {
	if err := w.doRequest(http.MethodGet, "/graph", nil); err != nil {
		return nil, err
	}
	var graph struct {
		Edges []map[string]any `json:"edges"`
	}
	if err := json.Unmarshal(w.lastBody, &graph); err != nil {
		return nil, fmt.Errorf("graph is not JSON: %v; body=%s", err, w.lastBody)
	}
	return graph.Edges, nil
}

func edgeMatches(e map[string]any, consumer, provider, subject string) bool {
	return fmt.Sprintf("%v", e["consumer"]) == consumer &&
		fmt.Sprintf("%v", e["provider"]) == provider &&
		fmt.Sprintf("%v", e["subject"]) == subject
}

// edgeState — найти ребро и проверить флаг (compatible / stale).
func (w *World) edgeState(consumer, provider, subject, flag string, want bool) error {
	edges, err := w.graphEdges()
	if err != nil {
		return err
	}
	for _, e := range edges {
		if !edgeMatches(e, consumer, provider, subject) {
			continue
		}
		got, _ := e[flag].(bool)
		if got != want {
			return fmt.Errorf("edge %s → %s %q: %s = %v, want %v (edge=%v)", consumer, provider, subject, flag, got, want, e)
		}
		return nil
	}
	return fmt.Errorf("edge %s → %s %q not found in graph", consumer, provider, subject)
}

// bodyNotContains — в теле последнего ответа нет подстроки.
func (w *World) bodyNotContains(sub string) error {
	if strings.Contains(string(w.lastBody), sub) {
		return fmt.Errorf("тело не должно содержать %q: %s", sub, w.lastBody)
	}
	return nil
}

// polygonSteps — регистрация шагов полигона (используются только polygon.feature).
func (w *World) polygonSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^бинарь pinout-openapi проверяет пару (\S+) и отчёт принят$`, func(p string) error {
		return w.runValidatorAndIngest("TOOL_OPENAPI", "/fixtures-poly/consumers/"+p+"/config.yaml")
	})
	ctx.Step(`^бинарь pinout-asyncapi проверяет пару (\S+) и отчёт принят$`, func(p string) error {
		return w.runValidatorAndIngest("TOOL_ASYNCAPI", p+"/config.yaml")
	})
	ctx.Step(`^ребро (\S+) → (\S+) "([^"]*)" совместимо$`, func(c, p, s string) error {
		return w.edgeState(c, p, s, "compatible", true)
	})
	ctx.Step(`^ребро (\S+) → (\S+) "([^"]*)" несовместимо$`, func(c, p, s string) error {
		return w.edgeState(c, p, s, "compatible", false)
	})
	ctx.Step(`^ребро (\S+) → (\S+) "([^"]*)" помечено stale$`, func(c, p, s string) error {
		return w.edgeState(c, p, s, "stale", true)
	})
	ctx.Step(`^ответ НЕ содержит значение "([^"]*)"$`, w.bodyNotContains)
}

// TestPolygon — отдельный сьют полигона (features/polygon.feature).
func TestPolygon(t *testing.T) {
	status := godog.TestSuite{
		Name:                "polygon",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) { newWorld().polygonSuite(ctx) },
		Options: &godog.Options{
			Paths:     []string{"../features-polygon/polygon.feature"},
			Format:    "pretty",
			Randomize: -1,
		},
	}.Run()
	if status != 0 {
		t.Fail()
	}
}

// polygonSuite — поли-сьюту нужны и базовые HTTP-шаги (статусы/поля/файлы).
func (w *World) polygonSuite(ctx *godog.ScenarioContext) {
	ctx.Before(w.beforeScenario)
	ctx.After(w.afterScenario)
	w.registerHTTPSteps(ctx)
	w.registerIngestSteps(ctx)
	w.registerImpactSteps(ctx)
	w.polygonSteps(ctx)
}
